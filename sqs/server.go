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
	Config *SqsConfig
	Store  storage.Store
}

// NewServer craetes a new SQS server backed by a given storage impl
func NewServer(config *SqsConfig, store storage.Store) *Server {
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
	switch action {
	case "CreateQueue":
		s.createQueue(c)
	case "DeleteMessage":
	case "DeleteQueue":
	//case "GetQueueAttributes":
	case "GetQueueUrl":
	case "ListQueues":
		s.listQueues(c)
	case "PurgeQueue":
	case "ReceiveMessage":
	case "SendMessage":
	// case "SetQueueAttributes":
	// TODO: batch stuff
	default:
		log.Printf("Unsupported Action : %s", action)
		c.String(http.StatusNotFound, "The requested resource could not be found.")
		return
	}
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
	if err != nil {
		c.XML(http.StatusInternalServerError, NewErrorResponse("Sender", err.Error()))
		return
	}
	for _, queue := range list.Queues {
		log.Printf("Queue: %s", queue)
	}
}

func (s *Server) sendBadRequest(c *gin.Context) {
	c.String(http.StatusBadRequest, fmt.Sprintf("Unsupported request %s", c.Request.URL.String()))
}
