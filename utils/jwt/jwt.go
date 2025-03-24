package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/handball0/tako/global/variable"
	"github.com/spf13/viper"
	"time"
)

// JwtInstance Jwt实例
type JwtInstance struct {
	secretKey []byte
}

// NewJwt 创建Jwt实例
func NewJwt() *JwtInstance {

	secretKey := []byte(viper.GetString("Jwt.secretKey"))
	if secretKey == nil || len(secretKey) == 0 {
		secretKey = []byte("AllYourBaseAreBelongToUs")
		variable.Echo.Warnf("Jwt secretKey is nil, use default value: %s\n", secretKey)
	}

	return &JwtInstance{
		secretKey: secretKey,
	}
}

// GenerateToken 生成Token
func (j *JwtInstance) GenerateToken(claimsMap jwt.MapClaims) (string, error) {

	variable.Echo.Infof("SecretKey: %s\n", j.secretKey)
	if claimsMap == nil {
		claimsMap = make(jwt.MapClaims)
	}

	claimsMap["iat"] = jwt.NewNumericDate(time.Now())
	claimsMap["nbf"] = jwt.NewNumericDate(time.Now())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsMap)
	return token.SignedString(j.secretKey)
}

// ParseToken 解析Token
func (j *JwtInstance) ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, err
	}

	return claims, nil
}
