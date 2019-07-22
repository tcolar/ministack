package sqs

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tcolar/ministack/storage"
)

// Server creates the SQS server impl
type Server struct {
	Config *Config
	Store  storage.Store
}

// NewServer craetes a new SQS server backed by a given storage impl
func NewServer(config *Config, store storage.Store) *Server {
	return &Server{
		Config: config,
		Store:  store,
	}
}

// Start sarts the server (runs forver)
func (s *Server) Start() {
	router := gin.Default()
	s.addRoutes(router)
	// router.Configure(lion.WithNotFoundHandler(notImplementedHandler))
	router.Run(fmt.Sprintf(":%d", s.Config.Port))
}

func (s *Server) addRoutes(router *gin.Engine) {
	router.GET("/", s.home)
	router.POST("/", s.home)
	router.NoRoute(s.sendBadRequest)
}

func (s *Server) home(c *gin.Context) {
	action := c.Query("Action")
	if len(action) == 0 {
		log.Printf("The Action parameter is missing in %s", c.Request.URL.String())
		s.sendBadRequest(c)
		return
	}
	// TODO: all the batch stuff
	switch action {
	case "AddPermission":
		s.AddPermissions(c)
	case "ChangeMessageVisibility":
	case "CreateQueue":
		s.createQueue(c)
	case "DeleteMessage":
	case "DeleteQueue": // 1
	case "GetQueueAttributes":
	case "GetQueueUrl": // 3
	case "ListDeadLetterSourceQueues":
	case "ListQueues":
		s.listQueues(c)
	case "ListQueueTags":
	case "PurgeQueue": // 2
	case "ReceiveMessage": // 5
	case "RemovePermission":
		s.RemovePermissions(c)
	case "SendMessage": // 4
	case "SetQueueAttributes":
	case "TagQueue":
	case "UntagQueue":
	default:
		log.Printf("Unsupported Action : %s", action)
		c.String(http.StatusNotFound, "The requested resource could not be found.")
		return
	}
}

func (s *Server) addPermissions(c *gin.Context) {
	log.Println("AddPermission is not implemented - Noop")
	c.XML(200, NewAddPermissionResponse())
}

func (s *Server) createQueue(c *gin.Context) {
	name := c.Query("QueueName")
	if len(name) == 0 {
		error := NewErrorResponse("Sender", "Invalid request: MissingQueryParamRejection(QueueName)")
		c.XML(http.StatusBadRequest, error)
		return
	}
	err := s.Store.CreateQueue(name)
	if err != nil {
		c.XML(http.StatusInternalServerError, NewErrorResponse("Sender", err.Error()))
		return
	}
	response := CreateQueueResponse{
		CreateQueueResult: CreateQueueResult{
			QueueUrl: fmt.Sprintf("http://%s:%d/queue/%s", s.Config.Host, s.Config.Port, name),
		},
		ResponseMetadata: ResponseMetadata{
			RequestId: DummyRequestID,
		},
	}
	c.XML(200, response)
}

func (s *Server) listQueues(c *gin.Context) {
	list, err := s.Store.ListQueues()
	// TODO: optional QueueNamePrefix arg
	if err != nil {
		c.XML(http.StatusInternalServerError, NewErrorResponse("Sender", err.Error()))
		return
	}
	response := NewListQueueResponse(s.Config, list.Keys())
	c.XML(200, response)
}

func (s *Server) removePermissions(c *gin.Context) {
	log.Println("RemovePermission is not implemented - Noop")
	c.XML(200, new RemovePermissionResponse())
}

func (s *Server) sendBadRequest(c *gin.Context) {
	c.String(http.StatusBadRequest, fmt.Sprintf("Unsupported request %s", c.Request.URL.String()))
}
