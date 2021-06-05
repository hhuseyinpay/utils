// custom context kullanmak için. endpoint sayısı artarsa custom .Get .Post vs methodları implemente edilebilir
// app.Get("/kullanici", CtxWrap(handlers.GetAll))
func CtxWrap(h func(ctx *context.AppCtx) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return h(&context.AppCtx{Ctx: c, Db: database.DB()})
	}
}

// limiter middleware
v1.Use(limiter.New(limiter.Config{
  Next: func(c *fiber.Ctx) bool {
    return c.IP() == "127.0.0.1"
  },
  Max:        10,
  Expiration: 30 * time.Second,
  KeyGenerator: func(c *fiber.Ctx) string {
    return strings.Split(c.Get("X-Forwarded-For", ","), ",")[0]
  },
  LimitReached: func(c *fiber.Ctx) error {
    return utils.ErrorBadRequest("çok fazla deneme yaptınız. lütfen daha sonra tekrar deneyiniz")
  },
}))

// app.Get("/health", health)
func health(c *fiber.Ctx) error {
	type Status struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	//var home model.Home
	//err := database.DB().First(&home).Error
	//if err != nil {
	//	return c.Status(400).JSON(Status{
	//		Status:  "error",
	//		Message: "database objesi alınamadı",
	//	})
	//}

	//err = cachesystem.Ping()
	//if err != nil {
	//	c.JSON(400, Status{
	//		Status:  "error",
	//		Message: "redis ping error",
	//	})
	//	return
	//}

	return c.JSON(Status{
		Status:  "OK",
		Message: "maşşşallah len",
	})
}
