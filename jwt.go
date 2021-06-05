
import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"strings"
)

func jwtErrorHandler(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{"status": "error", "message": "Missing or malformed JWT", "data": nil})
	} else {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{"status": "error", "message": "Invalid or expired JWT", "data": nil})
	}
}

func jwtSuccessHandler(c *fiber.Ctx) error {
	kullanici := model.Kullanici{}
	tokenByte := c.Request().Header.Peek("Authorization")
	if len(tokenByte) == 0 {
		return utils.ErrorNotLogged("Login Gerekli")
	}

	tokenStr := strings.ReplaceAll(string(c.Request().Header.Peek("Authorization")), "Bearer ", "")
	if tokenStr == "" {
		return utils.ErrorNotLogged("Login Gerekli")
	}

	jwtToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Get().Server.JwtSecret), nil
	})
	if err != nil {
		return utils.ErrorNotLogged(err.Error())
	}
	if !jwtToken.Valid {
		return utils.ErrorNotLogged(err.Error())
	}
	claims := jwtToken.Claims.(jwt.MapClaims)

	kullaniciId := claims["kullaniciID"].(float64)
	kullanici.ID = int64(kullaniciId)

	kullaniciYetki := claims["yetki"].(float64)
	kullanici.Yetki = model.KullaniciYetki(kullaniciYetki)

	kullaniciEPosta := claims["eposta"]
	kullanici.Eposta = fmt.Sprintf("%v", kullaniciEPosta)
	c.Locals("kullanici", kullanici)
	return c.Next()
}
