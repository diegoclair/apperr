package apperr

// Code is a unique and stable identifier for each system error.
// The frontend uses this code to translate messages (i18n).
// Convention: "DOMAIN_ACTION_REASON" (e.g., AUTH_LOGIN_BLOCKED, USER_EMAIL_EXISTS)
type Code string
