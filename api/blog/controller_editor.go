package blog

import (
	"github.com/gin-gonic/gin"
	userModel "github.com/unusualcodeorg/go-lang-backend-architecture/api/user/model"
	"github.com/unusualcodeorg/go-lang-backend-architecture/framework/network"
)

type editorController struct {
	network.BaseController
	service Service
}

func NewEditorController(
	authMFunc network.AuthenticationProvider,
	authorizeMFunc network.AuthorizationProvider,
	service Service,
) network.Controller {
	return &editorController{
		BaseController: network.NewBaseController("/blog/editor", authMFunc, authorizeMFunc),
		service:        service,
	}
}

func (c *editorController) MountRoutes(group *gin.RouterGroup) {
	group.Use(c.Authentication(), c.Authorization(string(userModel.RoleCodeWriter)))
	group.GET("/id/:id", c.getBlogHandler)

}

func (c *editorController) getBlogHandler(ctx *gin.Context) {

}
