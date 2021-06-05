package context

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/locales/tr"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	turkish "github.com/go-playground/validator/v10/translations/tr"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type ResponseModel struct {
	Data      interface{} `json:"data"`
	Message   string      `json:"message"`
	HataVarMi bool        `json:"hataVarMi"`
}

type PaginationModel struct {
	Offset    int    `query:"offset"`
	Page      int    `query:"page"`
	PerPage   int    `query:"perPage"`
	SortField string `query:"sortField"`
	SortType  string `query:"sortType"`
}

func (p PaginationModel) String() string {
	return strconv.Itoa(p.Page) + "-" + strconv.Itoa(p.PerPage)
}

type FilterModel struct {
	PaginationModel
	SearchFields []string `query:"searchFields"`
	SearchTerms  []string `query:"searchTerms"`
}

type AppCtx struct {
	*fiber.Ctx
	Db         *gorm.DB
	AuditModel *model.Audit
}

func (c *AppCtx) SuccessResponse(data interface{}) error {
	model := &ResponseModel{
		Data:      data,
		HataVarMi: false,
	}

	return c.JSON(model)
}

func (c *AppCtx) SuccessAndTotalRecordsResponse(data interface{}, totalRecords int64) error {
	result := fiber.Map{
		"data":         data,
		"totalRecords": totalRecords,
	}
	model := &ResponseModel{
		Data:      result,
		HataVarMi: false,
	}

	return c.JSON(model)
}

func (c *AppCtx) NotFoundResponse() error {
	model := &ResponseModel{
		HataVarMi: true,
		Message:   "Kayıt bulunamadı",
	}

	return c.Status(404).JSON(model)
}

func (c *AppCtx) ErrorResponse(code int, msg string) error {
	model := &ResponseModel{
		HataVarMi: true,
		Message:   msg,
	}

	return c.Status(code).JSON(model)
}

func (c *AppCtx) Log() *zap.Logger {
	ctxRqId := c.Get("requestid", "")
	return config.Logger(ctxRqId)
}

func (c *AppCtx) GetPaginationModel() (*PaginationModel, error) {
	model := new(PaginationModel)
	if err := c.QueryParser(model); err != nil {
		return nil, err
	}
	if model.Page == 0 {
		model.Page = 1
	}
	if model.PerPage == 0 {
		model.PerPage = 100
	}
	model.Offset = (model.Page - 1) * model.PerPage
	return model, nil
}

func (c *AppCtx) GetFilterModel() (*FilterModel, error) {
	model := new(FilterModel)
	if err := c.QueryParser(model); err != nil {
		return nil, err
	}

	if model.SortField == "" {
		model.SortField = "id"
	}
	if model.SortType == "" {
		model.SortType = "asc"
	}

	if model.Page == 0 {
		model.Page = 1
	}
	if model.PerPage == 0 {
		model.PerPage = 100
	}
	model.Offset = (model.Page - 1) * model.PerPage
	return model, nil
}

func (c *AppCtx) GetLang() string {
	l := c.Get("Accept-Language", "tr")
	if len(l) > 2 {
		return "tr"
	}
	return l
}

func (c *AppCtx) GetFromCache(key string, model interface{}) error {
	result, err := cache.Get(c.Context(), key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return err
		}
		_, file, no, ok := runtime.Caller(1)
		var filename string
		if ok {
			filename = fmt.Sprintf("called from %s#%d", file, no)
		}
		c.Log().Error("getcache from: "+filename, zap.Error(err))
		return err
	}

	err = json.Unmarshal([]byte(result), model)
	if err != nil {
		_, file, no, ok := runtime.Caller(1)
		var filename string
		if ok {
			filename = fmt.Sprintf("called from %s#%d", file, no)
		}
		c.Log().Error("getcache from: "+filename, zap.Error(err))
		return err
	}
	return nil
}

func (c *AppCtx) SetToCache(key string, model interface{}) {
	data, err := json.Marshal(model)
	if err != nil {
		_, file, no, ok := runtime.Caller(1)
		var filename string
		if ok {
			filename = fmt.Sprintf("called from %s#%d", file, no)
		}
		c.Log().Error("setcache from: "+filename, zap.Error(err))
	}

	err = cache.Set(c.Context(), key, data, 10*time.Second)
	if err != nil {
		_, file, no, ok := runtime.Caller(1)
		var filename string
		if ok {
			filename = fmt.Sprintf("called from %s#%d", file, no)
		}
		c.Log().Error("setcache from: "+filename, zap.Error(err))
	}
}

func (c *AppCtx) SetToCacheTTL(key string, model interface{}, duration time.Duration) {
	data, err := json.Marshal(model)
	if err != nil {
		_, file, no, ok := runtime.Caller(1)
		var filename string
		if ok {
			filename = fmt.Sprintf("called from %s#%d", file, no)
		}
		c.Log().Error("setcache from: "+filename, zap.Error(err))
	}

	err = cache.Set(c.Context(), key, data, duration)
	if err != nil {
		_, file, no, ok := runtime.Caller(1)
		var filename string
		if ok {
			filename = fmt.Sprintf("called from %s#%d", file, no)
		}
		c.Log().Error("setcache from: "+filename, zap.Error(err))
	}
}

func (c *AppCtx) BodyParserAndValidation(model interface{}) error {
	if err := c.BodyParser(&model); err != nil {
		return errors.New("hatalı model: " + err.Error())
	}

	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("labelName"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	tr := tr.New()
	uni := ut.New(tr, tr)
	trans, _ := uni.GetTranslator("tr")
	turkish.RegisterDefaultTranslations(validate, trans)

	err := validate.Struct(model)
	if err != nil {
		msg := ""
		errs := err.(validator.ValidationErrors)
		for _, e := range errs {
			msg += e.Translate(trans) + "\n"
		}
		return errors.New(msg)
	}
	return nil
}

func (c *AppCtx) GetUserID() int64 {
	user := c.Locals("kullanici")
	if user == nil {
		return 0
	}
	return user.(model.Kullanici).ID
}

func (c *AppCtx) GetUser() model.Kullanici {
	user := c.Locals("kullanici")
	if user == nil {
		return model.Kullanici{}
	}
	return user.(model.Kullanici)
}

func (c *AppCtx) GetUserJson() []byte {
	user := c.GetUser()
	if user.ID == 0 {
		return nil
	}
	userjson, _ := json.Marshal(c.GetUser())
	return userjson
}

func (c *AppCtx) InitAuditLogCreate(tabloismi string) {
	c.AuditModel = &model.Audit{
		CreatedAt:   time.Now(),
		KullaniciID: c.GetUserID(),
		Kullanici:   c.GetUserJson(),
		TabloIsmi:   tabloismi,
		Islem:       model.AuditIslemCreate,
		Durum:       model.AuditDurumError,
	}
	c.Locals("auditmodel", c.AuditModel)
}

func (c *AppCtx) InitAuditLogRead(tabloismi string) {
	c.AuditModel = &model.Audit{
		CreatedAt:   time.Now(),
		KullaniciID: c.GetUserID(),
		Kullanici:   c.GetUserJson(),
		TabloIsmi:   tabloismi,
		Islem:       model.AuditIslemRead,
		Durum:       model.AuditDurumError,
	}
	c.Locals("auditmodel", *c.AuditModel)
}

func (c *AppCtx) InitAuditLogUpdate(tabloismi string) {
	c.AuditModel = &model.Audit{
		CreatedAt:   time.Now(),
		KullaniciID: c.GetUserID(),
		Kullanici:   c.GetUserJson(),
		TabloIsmi:   tabloismi,
		Islem:       model.AuditIslemUpdate,
		Durum:       model.AuditDurumError,
	}
	c.Locals("auditmodel", *c.AuditModel)
}

func (c *AppCtx) InitAuditLogDelete(tabloismi string) {
	c.AuditModel = &model.Audit{
		CreatedAt:   time.Now(),
		KullaniciID: c.GetUserID(),
		Kullanici:   c.GetUserJson(),
		TabloIsmi:   tabloismi,
		Islem:       model.AuditIslemDelete,
		Durum:       model.AuditDurumError,
	}
	c.Locals("auditmodel", *c.AuditModel)
}

func (c *AppCtx) SetAuditLog(prevModel, newModel interface{}) {
	prevJson, _ := json.Marshal(prevModel)
	c.AuditModel.OncekiModel = prevJson

	newJson, _ := json.Marshal(newModel)
	c.AuditModel.OncekiModel = newJson
	c.AuditModel.Durum = model.AuditDurumSuccess
	c.Locals("auditmodel", *c.AuditModel)
}

func (c *AppCtx) GetIpAddress() string {
	return strings.Split(c.Get("X-Forwarded-For", ","), ",")[0]
}
