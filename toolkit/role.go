package toolkit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/session"
	engine "github.com/muidea/magicEngine"
)

// RoleVerifier role verifier
type RoleVerifier interface {
	CasVerifier
	VerifyRole(ctx context.Context, res http.ResponseWriter, req *http.Request) (*def.Role, error)
}

// RoleRegistry role route registry
type RoleRegistry interface {
	SetApiVersion(version string)

	AddHandler(pattern, method string, privateValue int, handler func(context.Context, http.ResponseWriter, *http.Request))

	AddRoute(route engine.Route, privateValue int, filters ...engine.MiddleWareHandler)

	GetAllPrivateItem() []*def.PrivateItem
}

// NewRoleRegistry create routeRegistry
func NewRoleRegistry(verifier RoleVerifier, router engine.Router) RoleRegistry {
	return &roleRegistryImpl{roleVerifier: verifier, router: router, privateItemSlice: privateItemSlice{}}
}

type privateItem struct {
	patternFilter *engine.PatternFilter
	privateValue  int
	patternPath   string
}

type privateItemSlice []*privateItem

// roleRegistryImpl cas route registry
type roleRegistryImpl struct {
	roleVerifier     RoleVerifier
	router           engine.Router
	privateItemSlice privateItemSlice
}

func (s *roleRegistryImpl) SetApiVersion(version string) {
	s.router.SetApiVersion(version)
}

// AddHandler add route handler
func (s *roleRegistryImpl) AddHandler(
	pattern, method string,
	privateValue int,
	handler func(context.Context, http.ResponseWriter, *http.Request)) {

	rtPattern := pattern
	apiVersion := s.router.GetApiVersion()
	if apiVersion != "" {
		rtPattern = fmt.Sprintf("%s%s", apiVersion, rtPattern)
	}

	privateItem := &privateItem{
		patternFilter: engine.NewPatternFilter(rtPattern),
		privateValue:  privateValue,
		patternPath:   rtPattern,
	}

	s.privateItemSlice = append(s.privateItemSlice, privateItem)

	s.router.AddRoute(engine.CreateRoute(pattern, method, handler), s)
}

func (s *roleRegistryImpl) AddRoute(route engine.Route, privateValue int, filters ...engine.MiddleWareHandler) {
	privateItem := &privateItem{
		patternFilter: engine.NewPatternFilter(route.Pattern()),
		privateValue:  privateValue,
		patternPath:   route.Pattern(),
	}

	s.privateItemSlice = append(s.privateItemSlice, privateItem)

	filters = append(filters, s)
	s.router.AddRoute(route, filters...)
}

func (s *roleRegistryImpl) GetAllPrivateItem() (ret []*def.PrivateItem) {
	for _, val := range s.privateItemSlice {
		item := &def.PrivateItem{Path: val.patternPath, Value: def.GetPrivateInfo(val.privateValue)}

		ret = append(ret, item)
	}

	return
}

// Handle middleware handler
func (s *roleRegistryImpl) Handle(ctx engine.RequestContext, res http.ResponseWriter, req *http.Request) {
	result := &def.Result{ErrorCode: def.Success}
	for {
		// must verify cas
		casEntity, casErr := s.roleVerifier.Verify(ctx.Context(), res, req)
		if casErr != nil {
			result.ErrorCode = def.InvalidAuthority
			result.Reason = casErr.Error()
			break
		}

		casCtx := context.WithValue(ctx.Context(), session.AuthAccount, casEntity)
		casRole, casErr := s.roleVerifier.VerifyRole(casCtx, res, req)
		if casErr != nil {
			result.ErrorCode = def.InvalidAuthority
			result.Reason = casErr.Error()
			break
		}

		privatePattern := ""
		privateValue := 0
		for _, val := range s.privateItemSlice {
			if val.patternFilter.Match(req.URL.Path) {
				privatePattern = val.patternPath
				privateValue = val.privateValue
				break
			}
		}

		err := s.verifyRole(casRole, privatePattern, privateValue)
		if err != nil {
			result.ErrorCode = def.InvalidAuthority
			result.Reason = err.Error()
			break
		}

		roleCtx := context.WithValue(casCtx, session.AuthRole, casRole)
		ctx.Update(roleCtx)
		break
	}

	if result.Fail() {
		block, err := json.Marshal(result)
		if err == nil {
			res.Write(block)
			return
		}

		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx.Next()
}

func (s *roleRegistryImpl) verifyRole(privateRole *def.Role, privatePath string, privateValue int) (err error) {
	var privateLite *def.PrivateItem
	//for {
	// 如果是处于初始化状态的administrator账号，则认为有权限(特殊判断)
	//if accountInfoVal.Account == "administrator" && accountInfoVal.Status.IsInitStatus() && accountInfoVal.Role == nil {
	//	return nil
	//}

	privateLite = s.checkPrivate(privatePath, privateRole)
	//	break
	//}
	if privateLite == nil {
		return fmt.Errorf("无效权限组")
	}

	if privateLite.Value.Value >= privateValue {
		return nil
	}

	return fmt.Errorf("当前账号无操作权限")
}

func (s *roleRegistryImpl) checkPrivate(privatePath string, privateRole *def.Role) (ret *def.PrivateItem) {
	if privateRole == nil {
		return
	}

	for ii := range privateRole.Private {
		val := privateRole.Private[ii]
		if val.Path == "*" {
			ret = val
			break
		}

		if val.Path == privatePath {
			ret = val
			break
		}
	}

	return
}
