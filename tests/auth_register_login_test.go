package tests

import (
	authv1 "Authorization-Service/contracts/gen/go/auth"
	"Authorization-Service/service/tests/suite"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	emptyAppID = 0
	appID      = 1
	appSecret  = "test-secret"

	passDefaultLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := RandomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respLogin, err := st.AuthClient.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appID,
	})
	require.NoError(t, err)

	loginTime := time.Now()

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"])
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	const DeltaSeconds = 5

	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), DeltaSeconds)

}

func TestRegisterLogin_Login_WrongPassword(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := RandomFakePassword()

	responseFromRegistration, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, responseFromRegistration.GetUserId())

	_, err = st.AuthClient.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password + "1",
		AppId:    appID,
	})
	require.Error(t, err)
}

func TestRegisterLogin_DuplicatedRegistration(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := RandomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respReg, err = st.AuthClient.Register(ctx, &authv1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

func TestRegisterLogin_Login_EmptyEmail(t *testing.T) {
	ctx, st := suite.New(t)

	_, err := st.AuthClient.Login(ctx, &authv1.LoginRequest{
		Email:    "",
		Password: RandomFakePassword(),
		AppId:    emptyAppID,
	})
	require.Error(t, err)
}

func TestRegisterLogin_Login_EmptyPassword(t *testing.T) {
	ctx, st := suite.New(t)

	_, err := st.AuthClient.Login(ctx, &authv1.LoginRequest{
		Email:    gofakeit.Email(),
		Password: "",
		AppId:    emptyAppID,
	})
	require.Error(t, err)
}

func TestRegisterLogin_Login_EmptyAppID(t *testing.T) {
	ctx, st := suite.New(t)

	_, err := st.AuthClient.Login(ctx, &authv1.LoginRequest{
		Email:    gofakeit.Email(),
		Password: RandomFakePassword(),
		AppId:    emptyAppID,
	})
	require.Error(t, err)
}

func RandomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}
