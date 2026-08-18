package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/avast/retry-go/v5"
	"github.com/mastastny/slavoj-web-2025/internal/config"
	"github.com/mastastny/slavoj-web-2025/internal/repository"
)

type LockerService struct {
	conf      config.Config
	client    http.Client
	tokenRepo repository.LockerTokenRepository // nil-able: no cross-cold-start persistence when nil

	tokenMu       sync.Mutex
	tokenIssuedAt time.Time // zero value means "age unknown", forces a refresh
}

// NewLockerService kicks off loading any persisted token from tokenRepo
// (which may be nil, e.g. when Supabase isn't configured) in the background,
// so the DB round trip doesn't add to Lambda init/cold-start latency for
// requests that never touch the locker. If a request needs a token before
// the load finishes, getAccessToken simply blocks on the same mutex - no
// worse than doing the load synchronously here, just not on the critical
// path when nothing needs it yet.
func NewLockerService(conf config.Config, tokenRepo repository.LockerTokenRepository) *LockerService {
	l := &LockerService{
		conf:      conf,
		client:    http.Client{},
		tokenRepo: tokenRepo,
	}
	if tokenRepo != nil {
		go l.loadPersistedToken()
	}
	return l
}

func (l *LockerService) loadPersistedToken() {
	l.tokenMu.Lock()
	defer l.tokenMu.Unlock()

	// getAccessToken may have already run (and refreshed) while this
	// goroutine was waiting for the lock - in that case the in-memory state
	// is newer than whatever this load would find, so don't clobber it.
	if !l.tokenIssuedAt.IsZero() {
		return
	}

	stored, found, err := l.tokenRepo.LoadLockerToken()
	if err != nil {
		slog.Warn("loadPersistedToken -> cannot load persisted locker token, falling back to configured token", "err", err)
		return
	}
	if found {
		l.conf.LockerService.AccessToken = stored.AccessToken
		l.conf.LockerService.RefreshToken = stored.RefreshToken
		l.conf.LockerService.TokenLifespan = stored.TokenLifespan
		l.tokenIssuedAt = stored.IssuedAt
	}
}

// apiErrorResponse is the common error envelope used by the Sciener/TTLock
// API. It still returns HTTP 200 on failure, so every response body must be
// checked for a non-zero errcode.
type apiErrorResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func checkAPIError(body []byte) error {
	var apiErr apiErrorResponse
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return nil
	}
	if apiErr.ErrCode != 0 {
		return fmt.Errorf("api error %d: %s", apiErr.ErrCode, apiErr.ErrMsg)
	}
	return nil
}

// todo stara funkce bude smazana
func (l *LockerService) GetLockerCode(targetTime time.Time) string {
	if l.conf.Areal.CabinetLockCode != "" {
		return l.conf.Areal.CabinetLockCode
	}

	tz, err := time.LoadLocation("Europe/Prague")
	if err != nil { // Always check errors even if they should not happen.
		panic(err)
	}

	return targetTime.In(tz).Format("0106")
}

// CreatePasscode creates new passcode for the given time and sends it to the locker
//
// passcode will be valid 30 minutes before and 30 minutes after the event
func (l *LockerService) CreatePasscode(startTime time.Time, endTime time.Time) string {
	lockerStart := startTime.Add(-30 * time.Minute)
	lockerEnd := endTime.Add(30 * time.Minute)
	passcodeCandidate := ""
	err := retry.New( // todo use RetryIf and retry ony on code collision
		retry.Attempts(3),
		retry.Delay(100*time.Millisecond),
	).Do(
		func() error {
			passcodeCandidate = fmt.Sprintf("%06d", rand.Intn(1000000))
			err := l.sendPasscodeToLock(lockerStart, lockerEnd, passcodeCandidate, passcodeCandidate)
			if err != nil {
				slog.Warn("CreatePasscode -> sendPasscodeToLock failed", "err", err)
				return err
			}
			slog.Info("CreatePasscode -> new passcode was uploaded to locker", "passcode", passcodeCandidate)
			return nil
		},
	)
	if err != nil {
		slog.Warn("CreatePasscode -> cannot create new passcode, using backup code")
		return l.conf.LockerService.BackupCode
	}
	return passcodeCandidate
}

type passcodeResponse struct {
	KeyboardPwdID string `json:"keyboard_pwd_id"`
}

// sendPasscodeToLock sends new passcode to the locker
//
// StartTime and endTime are real times when the passcode should be valid.
// Pwd must be unique.
func (l *LockerService) sendPasscodeToLock(startTime time.Time, endTime time.Time, pwd string, pwdName string) error {
	token, err := l.getAccessToken()
	if err != nil {
		return fmt.Errorf("sendPasscodeToLock -> cannot get access token: %w", err)
	}
	values := url.Values{
		"clientId":        {l.conf.LockerService.ClientId},
		"accessToken":     {token},
		"lockId":          {l.conf.LockerService.LockId},
		"keyboardPwd":     {pwd},
		"keyboardPwdName": {pwdName},
		"startDate":       {strconv.FormatInt(startTime.UnixMilli(), 10)},
		"endDate":         {strconv.FormatInt(endTime.UnixMilli(), 10)},
		"addType":         {"2"},
		"date":            {strconv.FormatInt(time.Now().UnixMilli(), 10)},
	}
	result, err := l.client.PostForm("https://api.sciener.com/v3/keyboardPwd/add", values)
	if err != nil {
		return fmt.Errorf("sendPasscodeToLock -> cannot send request: %w", err)
	}
	body, _ := io.ReadAll(result.Body)
	err = result.Body.Close()
	if err != nil {
		return fmt.Errorf("sendPasscodeToLock -> cannot close response body: %w", err)
	}
	slog.Debug("sendPasscodeToLock -> response", "status", result.StatusCode, "body", string(body))
	if apiErr := checkAPIError(body); apiErr != nil {
		// The cached token may be the reason the lock rejected the request;
		// invalidate it so the next retry attempt fetches a fresh one.
		l.invalidateAccessToken()
		return fmt.Errorf("sendPasscodeToLock -> %w", apiErr)
	}
	return nil
}

func (l *LockerService) getAccessToken() (string, error) {
	l.tokenMu.Lock()
	defer l.tokenMu.Unlock()

	lifespanDuration := time.Duration(l.conf.LockerService.TokenLifespan) * time.Second
	refreshDuration := time.Duration(l.conf.LockerService.RefreshWindow) * time.Second
	stablePeriod := lifespanDuration - refreshDuration

	// tokenIssuedAt is zero whenever we don't actually know the age of the
	// current token (fresh process, invalidated after an API error, or a
	// LOCKER_ACCESS_TOKEN preloaded from config of unknown age) - refresh
	// unconditionally in that case rather than trusting it.
	needsRefresh := l.conf.LockerService.AccessToken == "" ||
		l.tokenIssuedAt.IsZero() ||
		time.Since(l.tokenIssuedAt) > stablePeriod

	if needsRefresh {
		if err := l.updateAccessTokenLocked(); err != nil {
			return "", fmt.Errorf("getAccessToken -> refreshing access token failed: %w", err)
		}
	}
	return l.conf.LockerService.AccessToken, nil
}

// invalidateAccessToken marks the cached access token as stale, forcing the
// next getAccessToken call to fetch a new one.
func (l *LockerService) invalidateAccessToken() {
	l.tokenMu.Lock()
	l.tokenIssuedAt = time.Time{}
	l.tokenMu.Unlock()
}

// updateAccessTokenLocked must be called with tokenMu held.
func (l *LockerService) updateAccessTokenLocked() error {
	response, err := l.requestAccessToken()
	if err != nil {
		return fmt.Errorf("updateAccessTokenLocked -> request new access token failed: %w", err)
	}
	l.conf.LockerService.AccessToken = response.AccessToken
	l.conf.LockerService.RefreshToken = response.RefreshToken
	l.conf.LockerService.TokenLifespan = response.ExpiresIn // in seconds
	l.tokenIssuedAt = time.Now()

	if l.tokenRepo != nil {
		err := l.tokenRepo.SaveLockerToken(repository.LockerToken{
			AccessToken:   l.conf.LockerService.AccessToken,
			RefreshToken:  l.conf.LockerService.RefreshToken,
			TokenLifespan: l.conf.LockerService.TokenLifespan,
			IssuedAt:      l.tokenIssuedAt,
		})
		if err != nil {
			// Best effort: this container can still use the token it just
			// got, but the next cold start may race with a rotated refresh
			// token if this keeps failing.
			slog.Warn("updateAccessTokenLocked -> cannot persist locker token", "err", err)
		}
	}
	return nil
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (l *LockerService) requestAccessToken() (tokenResponse, error) {
	values := url.Values{
		"clientId":      {l.conf.LockerService.ClientId},
		"clientSecret":  {l.conf.LockerService.ClientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {l.conf.LockerService.RefreshToken},
	}
	resp, err := l.client.PostForm("https://euapi.ttlock.com/oauth2/token", values)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("requestAccessToken -> client post: %w", err)
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			slog.Warn("requestAccessToken -> cannot close response body", "err", err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("requestAccessToken -> read response: %w", err)
	}
	if apiErr := checkAPIError(body); apiErr != nil {
		return tokenResponse{}, fmt.Errorf("requestAccessToken -> %w", apiErr)
	}
	var result tokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return tokenResponse{}, fmt.Errorf("requestAccessToken -> decode response: %w", err)
	}
	return result, nil
}

type passcode struct {
	KeyboardPwdId      int    `json:"keyboardPwdId"`
	LockId             int    `json:"lockId"`
	KeyboardPwd        string `json:"keyboardPwd"`
	KeyboardPwdName    string `json:"keyboardPwdName"`
	KeyboardPwdVersion int    `json:"keyboardPwdVersion"`
	KeyboardPwdType    int    `json:"keyboardPwdType"`
	StartDate          int64  `json:"startDate"`
	EndDate            int64  `json:"endDate"`
	SendDate           int64  `json:"sendDate"`
	IsCustom           int    `json:"isCustom"`
	Status             int    `json:"status"`
	SenderUsername     string `json:"senderUsername"`
}

type listKeyboardPwdResponse struct {
	List []passcode `json:"list"`
}

func (l *LockerService) FilterExpiredPasscodesFromLock() error {
	passcodes, err := l.getAllPasscodesFromLock()
	if err != nil {
		return fmt.Errorf("filterExpiredPasscodesFromLock -> cannot get passcodes from lock: %w", err)
	}
	now := time.Now().UnixMilli()
	for _, passcode := range passcodes {
		expired := passcode.EndDate < now
		invalid := passcode.Status != 1
		if expired || invalid {
			err = l.deletePasscodeFromLock(passcode.KeyboardPwdId)
			if err != nil {
				slog.Warn("filterExpiredPasscodesFromLock -> cannot delete passcode", "keyboardPwdId", passcode.KeyboardPwdId)
			}
		}
	}
	return nil
}

func (l *LockerService) getAllPasscodesFromLock() ([]passcode, error) {
	token, err := l.getAccessToken()
	if err != nil {
		return nil, fmt.Errorf("getAllPasscodesFromLock -> cannot get access token: %w", err)
	}

	var all []passcode
	pageNo := 1
	for {
		values := url.Values{
			"clientId":    {l.conf.LockerService.ClientId},
			"accessToken": {token},
			"lockId":      {l.conf.LockerService.LockId},
			"pageNo":      {strconv.Itoa(pageNo)},
			"pageSize":    {"100"},
			"date":        {strconv.FormatInt(time.Now().UnixMilli(), 10)},
		}
		resp, err := l.client.PostForm("https://api.sciener.com/v3/lock/listKeyboardPwd", values)
		if err != nil {
			return nil, fmt.Errorf("getAllPasscodesFromLock -> cannot send request: %w", err)
		}
		body, _ := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if closeErr != nil {
			slog.Warn("getAllPasscodesFromLock -> cannot close response body", "err", closeErr)
		}
		if apiErr := checkAPIError(body); apiErr != nil {
			return nil, fmt.Errorf("getAllPasscodesFromLock -> %w", apiErr)
		}
		var result listKeyboardPwdResponse
		decodeErr := json.Unmarshal(body, &result)
		if decodeErr != nil {
			return nil, fmt.Errorf("getAllPasscodesFromLock -> decode response: %w", decodeErr)
		}
		slog.Debug("getAllPasscodesFromLock -> parsed", "count", len(result.List))

		all = append(all, result.List...)
		if len(result.List) < 100 {
			break
		}
		pageNo++
	}
	return all, nil
}

func (l *LockerService) deletePasscodeFromLock(keyboardPwdId int) error {
	token, err := l.getAccessToken()
	if err != nil {
		return fmt.Errorf("deletePasscodeFromLock -> cannot get access token: %w", err)
	}
	values := url.Values{
		"clientId":      {l.conf.LockerService.ClientId},
		"accessToken":   {token},
		"lockId":        {l.conf.LockerService.LockId},
		"keyboardPwdId": {strconv.Itoa(keyboardPwdId)},
		"deleteType":    {"2"},
		"date":          {strconv.FormatInt(time.Now().UnixMilli(), 10)},
	}
	response, err := l.client.PostForm("https://api.sciener.com/v3/keyboardPwd/delete", values)
	if err != nil {
		return fmt.Errorf("deletePasscodeFromLock -> cannot send request: %w", err)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Warn("deletePasscodeFromLock -> cannot read response body", "err", err)
	}
	err = response.Body.Close()
	if err != nil {
		slog.Warn("deletePasscodeFromLock -> cannot close response body", "err", err)
	}
	slog.Debug("deletePasscodeFromLock -> response", "status", response.StatusCode, "body", string(body))
	if apiErr := checkAPIError(body); apiErr != nil {
		return fmt.Errorf("deletePasscodeFromLock -> %w", apiErr)
	}
	slog.Info("deletePasscodeFromLock -> passcode deleted from lock", "pwdId", keyboardPwdId)
	return nil
}
