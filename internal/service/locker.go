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
	"time"

	"github.com/avast/retry-go/v5"
	"github.com/mastastny/slavoj-web-2025/internal/config"
)

type LockerService struct {
	conf         config.Config
	client       http.Client
	serviceStart time.Time
}

func NewLockerService(conf config.Config) *LockerService {
	return &LockerService{
		conf:         conf,
		client:       http.Client{},
		serviceStart: time.Now(),
	}
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
	return nil
}

func (l *LockerService) getAccessToken() (string, error) {
	if l.conf.LockerService.AccessToken == "" {
		err := l.updateAccessToken()
		if err != nil {
			return "", fmt.Errorf("getAccessToken -> getting first access token failed: %w", err)
		}
		return l.conf.LockerService.AccessToken, nil
	}
	lifespanDuration := time.Duration(l.conf.LockerService.TokenLifespan) * time.Second
	refreshDuration := time.Duration(l.conf.LockerService.RefreshWindow) * time.Second
	stablePeriod := lifespanDuration - refreshDuration
	if (time.Since(l.serviceStart)) > stablePeriod {
		err := l.updateAccessToken()
		if err != nil {
			return "", fmt.Errorf("getAccessToken -> getting first access token failed: %w", err)
		}
	} // todo bug, what if this function won't be called in refresh window?
	return l.conf.LockerService.AccessToken, nil
}

func (l *LockerService) updateAccessToken() error {
	response, err := l.requestAccessToken()
	if err != nil {
		return fmt.Errorf("updateAccessToken -> request new access token failed: %w", err)
	}
	l.conf.LockerService.AccessToken = response.AccessToken
	l.conf.LockerService.RefreshToken = response.RefreshToken
	l.conf.LockerService.TokenLifespan = response.ExpiresIn // in seconds
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
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	var result tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
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
	slog.Info("deletePasscodeFromLock -> passcode deleted from lock", "pwdId", keyboardPwdId)
	body, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Warn("deletePasscodeFromLock -> cannot read response body", "err", err)
	}
	err = response.Body.Close()
	if err != nil {
		slog.Warn("deletePasscodeFromLock -> cannot close response body", "err", err)
	}
	slog.Debug("deletePasscodeFromLock -> response", "status", response.StatusCode, "body", string(body))
	return nil
}
