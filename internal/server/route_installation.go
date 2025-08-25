package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/Nidal-Bakir/go-semver"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth"
	"github.com/Nidal-Bakir/go-todo-backend/internal/middleware"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"golang.org/x/text/language"
)

func installationRouter(_ context.Context, authRepo auth.Repository) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /create", createInstallation(authRepo))
	mux.HandleFunc("POST /update",
		middleware.MiddlewareChain(
			updateInstallation(authRepo),
			Installation(authRepo),
		),
	)

	return middleware.MiddlewareChain(
		mux.ServeHTTP,
		middleware.ACT_app_x_www_form_urlencoded,
	)
}

func createInstallation(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		params, errs := validateCreateInstallationParams(r)
		if len(errs) != 0 {
			writeError(ctx, w, r, http.StatusBadRequest, errs...)
			return
		}

		installationToken, err := authRepo.CreateInstallation(ctx, params)
		if err != nil {
			writeError(ctx, w, r, return400IfAppErrOr500(err), err)
			return
		}

		if params.ClientType.IsWeb() {
			setInstallationCookie(w, installationToken)
		}

		writeResponse(ctx, w, r, http.StatusCreated, map[string]string{"installation_token": installationToken})
	}
}

func validateCreateInstallationParams(r *http.Request) (param auth.CreateInstallationData, errs []error) {
	errs = make([]error, 0, 10)
	err := r.ParseForm()
	if err != nil {
		errs = append(errs, err)
		return param, errs
	}

	param.NotificationToken = r.FormValue("notification_token")

	param.Locale, err = parseLocale(r.FormValue("locale"))
	if err != nil {
		errs = append(errs, err)
	}

	param.TimezoneOffsetInMinutes, err = parseTimeZoneInMinutes(r.FormValue("timezone_offset_in_minutes"))
	if err != nil {
		errs = append(errs, errors.New("invalid timezone offset"))
	}

	deviceManufacturer := r.FormValue("device_manufacturer")
	if len(deviceManufacturer) < 50 {
		param.DeviceManufacturer = deviceManufacturer
	} else {
		errs = append(errs, errors.New("too long device manufacturer"))
	}

	deviceOS, err := new(auth.DeviceOS).FromString(r.FormValue("device_os"))
	if err != nil {
		errs = append(errs, err)
	} else {
		param.DeviceOS = *deviceOS
	}

	clientType, err := new(auth.ClientType).FromString(r.FormValue("client_type"))
	if err != nil {
		errs = append(errs, err)
	} else {
		param.ClientType = *clientType
	}

	deviceOsVersion := r.FormValue("device_os_version")
	if len(deviceOsVersion) < 10 {
		param.DeviceOSVersion = deviceOsVersion
	} else {
		errs = append(errs, errors.New("too long device OS version"))
	}

	param.AppVersion, err = parseAppVersion(r.FormValue("app_version"))
	if err != nil {
		errs = append(errs, err)
	}

	return param, errs
}

func updateInstallation(authRepo auth.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		params, errs := validateUpdateInstallationParams(r)
		if len(errs) != 0 {
			writeError(ctx, w, r, http.StatusBadRequest, errs...)
			return
		}

		installation, ok := auth.InstallationFromContext(ctx)
		utils.Assert(ok, "we shuld find the installation in the context but we did not. something is wrong")

		err := authRepo.UpdateInstallation(ctx, installation.InstallationToken, params)
		if err != nil {
			writeError(ctx, w, r, return400IfAppErrOr500(err), err)
			return
		}

		apiWriteOperationDoneSuccessfullyJson(ctx, w, r)
	}
}

func validateUpdateInstallationParams(r *http.Request) (param auth.UpdateInstallationData, errs []error) {
	errs = make([]error, 0, 3)
	err := r.ParseForm()
	if err != nil {
		errs = append(errs, err)
		return param, errs
	}

	param.AppVersion, err = parseAppVersion(r.FormValue("app_version"))
	if err != nil {
		errs = append(errs, err)
	}

	param.TimezoneOffsetInMinutes, err = parseTimeZoneInMinutes(r.FormValue("timezone_offset_in_minutes"))
	if err != nil {
		errs = append(errs, errors.New("invalid timezone offset"))
	}

	param.Locale, err = parseLocale(r.FormValue("locale"))
	if err != nil {
		errs = append(errs, err)
	}

	param.NotificationToken = r.FormValue("notification_token")

	return param, errs
}

func parseTimeZoneInMinutes(timezoneOffsetInMinutesStr string) (int, error) {
	t, err := strconv.Atoi(timezoneOffsetInMinutesStr)
	if err != nil {
		return 0, err
	} else {
		minDifferenceUTC := -12 * 60 // UTC−12:00
		maxDifferenceUTC := +14 * 60 // UTC+14:00
		if t < minDifferenceUTC || t > maxDifferenceUTC {
			return 0, errors.New("invalid timezone offset")
		}
	}
	return t, nil
}

func parseLocale(localeStr string) (string, error) {
	tag, err := language.Parse(localeStr)
	if err != nil {
		return "", errors.New("invalid locale")
	}
	return tag.String(), nil
}

func parseAppVersion(appVersionStr string) (string, error) {
	if semver.IsValid(appVersionStr) {
		return appVersionStr, nil
	}
	return "", errors.New("invalid app version. it should in the form x.y.z")
}
