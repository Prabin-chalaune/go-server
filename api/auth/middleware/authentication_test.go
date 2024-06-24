package middleware

import (
	"errors"
	"net/http"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unusualcodeorg/goserve/api/auth"
	"github.com/unusualcodeorg/goserve/api/auth/model"
	"github.com/unusualcodeorg/goserve/api/user"
	userModel "github.com/unusualcodeorg/goserve/api/user/model"
	"github.com/unusualcodeorg/goserve/arch/network"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAuthenticationProvider_NoAccessToken(t *testing.T) {
	mockAuthService := new(auth.MockService)
	mockUserService := new(user.MockService)

	mockAuthService.AssertNotCalled(t, "VerifyToken", mock.Anything)

	rr := network.MockTestAuthenticationProvider(t, "GET", "/test", "/test", "",
		NewAuthenticationProvider(mockAuthService, mockUserService),
		network.MockSuccessMsgHandler("success"),
	)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), `"message":"permission denied: missing Authorization"`)
	mockAuthService.AssertExpectations(t)
}

func TestAuthenticationProvider_WrongAccessToken(t *testing.T) {
	mockAuthService := new(auth.MockService)
	mockUserService := new(user.MockService)

	mockAuthService.AssertNotCalled(t, "VerifyToken", mock.Anything)

	token := "token"

	rr := network.MockTestAuthenticationProvider(t, "GET", "/test", "/test", "",
		NewAuthenticationProvider(mockAuthService, mockUserService),
		network.MockSuccessMsgHandler("success"),
		primitive.E{Key: network.AuthorizationHeader, Value: token},
	)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), `"message":"permission denied: invalid Authorization"`)
	mockAuthService.AssertExpectations(t)
}

func TestAuthenticationProvider_VerifyTokenInvalid(t *testing.T) {
	mockAuthService := new(auth.MockService)
	mockUserService := new(user.MockService)
	mockAuthService.AssertNotCalled(t, "ValidateClaims", mock.Anything)

	token := "Bearer token"

	mockAuthService.On("VerifyToken", "token").Return(nil, errors.New("invalid token"))

	rr := network.MockTestAuthenticationProvider(t, "GET", "/test", "/test", "",
		NewAuthenticationProvider(mockAuthService, mockUserService),
		network.MockSuccessMsgHandler("success"),
		primitive.E{Key: network.AuthorizationHeader, Value: token},
	)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), `"message":"invalid token"`)
	mockAuthService.AssertExpectations(t)
}

func TestAuthenticationProvider_VerifyTokenInvalidClaim(t *testing.T) {
	mockAuthService := new(auth.MockService)
	mockUserService := new(user.MockService)
	mockAuthService.AssertNotCalled(t, "FindUserById", mock.Anything)

	token := "Bearer token"
	claims := &jwt.RegisteredClaims{}

	mockAuthService.On("VerifyToken", "token").Return(claims, nil)
	mockAuthService.On("ValidateClaims", claims).Return(false)

	rr := network.MockTestAuthenticationProvider(t, "GET", "/test", "/test", "",
		NewAuthenticationProvider(mockAuthService, mockUserService),
		network.MockSuccessMsgHandler("success"),
		primitive.E{Key: network.AuthorizationHeader, Value: token},
	)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), `"message":"permission denied: invalid claims"`)
	mockAuthService.AssertExpectations(t)
}

func TestAuthenticationProvider_VerifyTokenInvalidClaimUser(t *testing.T) {
	mockAuthService := new(auth.MockService)
	mockUserService := new(user.MockService)
	mockAuthService.AssertNotCalled(t, "FindUserById", mock.Anything)

	token := "Bearer token"
	claims := &jwt.RegisteredClaims{}

	mockAuthService.On("VerifyToken", "token").Return(claims, nil)
	mockAuthService.On("ValidateClaims", claims).Return(true)

	rr := network.MockTestAuthenticationProvider(t, "GET", "/test", "/test", "",
		NewAuthenticationProvider(mockAuthService, mockUserService),
		network.MockSuccessMsgHandler("success"),
		primitive.E{Key: network.AuthorizationHeader, Value: token},
	)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), `"message":"permission denied: invalid claims subject"`)
	mockAuthService.AssertExpectations(t)
}

func TestAuthenticationProvider_VerifyTokenInvalidUser(t *testing.T) {
	mockAuthService := new(auth.MockService)
	mockUserService := new(user.MockService)
	mockAuthService.AssertNotCalled(t, "FindKeystore", mock.Anything)

	token := "Bearer token"
	userId := primitive.NewObjectID()
	claims := &jwt.RegisteredClaims{Subject: userId.Hex()}

	mockAuthService.On("VerifyToken", "token").Return(claims, nil)
	mockAuthService.On("ValidateClaims", claims).Return(true)
	mockUserService.On("FindUserById", userId).Return(nil, errors.New("user not found"))

	rr := network.MockTestAuthenticationProvider(t, "GET", "/test", "/test", "",
		NewAuthenticationProvider(mockAuthService, mockUserService),
		network.MockSuccessMsgHandler("success"),
		primitive.E{Key: network.AuthorizationHeader, Value: token},
	)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), `"message":"permission denied: claims subject does not exists"`)
	mockAuthService.AssertExpectations(t)
}

func TestAuthenticationProvider_VerifyTokenInvalidKaystore(t *testing.T) {
	mockAuthService := new(auth.MockService)
	mockUserService := new(user.MockService)

	token := "Bearer token"
	userId := primitive.NewObjectID()
	claims := &jwt.RegisteredClaims{ID: "claimId", Subject: userId.Hex()}
	user := &userModel.User{ID: userId}

	mockAuthService.On("VerifyToken", "token").Return(claims, nil)
	mockAuthService.On("ValidateClaims", claims).Return(true)
	mockUserService.On("FindUserById", userId).Return(user, nil)
	mockAuthService.On("FindKeystore", user, claims.ID).Return(nil, errors.New("not found"))

	rr := network.MockTestAuthenticationProvider(t, "GET", "/test", "/test", "",
		NewAuthenticationProvider(mockAuthService, mockUserService),
		network.MockSuccessMsgHandler("success"),
		primitive.E{Key: network.AuthorizationHeader, Value: token},
	)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), `"message":"permission denied: invalid access token"`)
}

func TestAuthenticationProvider_Success(t *testing.T) {
	mockAuthService := new(auth.MockService)
	mockUserService := new(user.MockService)

	token := "Bearer token"
	userId := primitive.NewObjectID()
	claims := &jwt.RegisteredClaims{ID: "claimId", Subject: userId.Hex()}
	user := &userModel.User{ID: userId}
	keystore := &model.Keystore{}

	mockAuthService.On("VerifyToken", "token").Return(claims, nil)
	mockAuthService.On("ValidateClaims", claims).Return(true)
	mockUserService.On("FindUserById", userId).Return(user, nil)
	mockAuthService.On("FindKeystore", user, claims.ID).Return(keystore, nil)

	rr := network.MockTestAuthenticationProvider(t, "GET", "/test", "/test", "",
		NewAuthenticationProvider(mockAuthService, mockUserService),
		network.MockSuccessMsgHandler("success"),
		primitive.E{Key: network.AuthorizationHeader, Value: token},
	)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"message":"success"`)
}
