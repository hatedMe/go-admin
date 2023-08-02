package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

// 自定义格式内容
type CustomerClaims struct {
	UserId         int    `json:"userId"`
	UserName       string `json:"userName"`
	RoleId         int    `json:"roleId"`
	StandardClaims jwt.StandardClaims
}

func (c CustomerClaims) Valid() error {
	return nil
}

// 生成token
func GenerateJwtToken(secret string, issuer string, audience string, expiredMinutes int64, userId int, userName string, roleId int) (string, error) {
	hmacSampleSecret := []byte(secret) //密钥，不能泄露
	token := jwt.New(jwt.SigningMethodHS256)
	nowTime := time.Now().Unix()
	token.Claims = CustomerClaims{
		UserName: userName,
		UserId:   userId,
		RoleId:   roleId,
		StandardClaims: jwt.StandardClaims{
			NotBefore: nowTime,                  // 签名生效时间
			ExpiresAt: nowTime + expiredMinutes, // 签名过期时间
			Issuer:    issuer,                   // 签名颁发者
			Audience:  audience,
		},
	}
	tokenString, err := token.SignedString(hmacSampleSecret)
	return tokenString, err
}

// 解析token
func ParseJwtToken(tokenString string, secret string) (*CustomerClaims, error) {
	var hmacSampleSecret = []byte(secret)
	//前面例子生成的token
	token, err := jwt.ParseWithClaims(tokenString, &CustomerClaims{}, func(t *jwt.Token) (interface{}, error) {
		return hmacSampleSecret, nil
	})

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	claims := token.Claims.(*CustomerClaims)
	return claims, nil
}
