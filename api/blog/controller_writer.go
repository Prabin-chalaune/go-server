package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/unusualcodeorg/go-lang-backend-architecture/api/blog/dto"
	userModel "github.com/unusualcodeorg/go-lang-backend-architecture/api/user/model"
	"github.com/unusualcodeorg/go-lang-backend-architecture/core/network"
)

type writerController struct {
	network.BaseController
	service Service
}

func NewWriterController(
	authMFunc network.AuthenticationProvider,
	authorizeMFunc network.AuthorizationProvider,
	service Service,
) network.Controller {
	return &writerController{
		BaseController: network.NewBaseController("/blog/writer", authMFunc, authorizeMFunc),
		service:    service,
	}
}

func (c *writerController) MountRoutes(group *gin.RouterGroup) {
	group.Use(c.Authentication(), c.Authorization(string(userModel.RoleCodeWriter)))
	group.POST("/", c.postBlogHandler)
}

func (c *writerController) postBlogHandler(ctx *gin.Context) {
	body, err := network.ReqBody(ctx, dto.EmptyCreateBlog())
	if err != nil {
		c.Send(ctx).BadRequestError(err.Error(), err)
		return
	}

	user := network.ReqMustGetUser[userModel.User](ctx)

	b, err := c.service.CreateBlog(body, user)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("blog creation success", b)
}