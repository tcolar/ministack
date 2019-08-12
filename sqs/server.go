package sqs

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicode"

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
		s.addPermissions(c)
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
		s.removePermissions(c)
	case "SendMessage": // 4
		s.sendMessage(c)
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
	if err := s.validateQueuName(name); err != nil {
		error := NewErrorResponse("Sender", fmt.Sprintf("Invalid queue name: %s", err))
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
	prefix := c.Query("QueueNamePrefix")
	list, err := s.Store.ListQueues()
	if err != nil {
		c.XML(http.StatusInternalServerError, NewErrorResponse("Sender", err.Error()))
		return
	}
	queues := list.Keys()
	var filteredList []string
	if len(prefix) == 0 {
		filteredList = queues
	} else {
		for _, q := range queues {
			if strings.HasPrefix(q, prefix) {
				filteredList = append(filteredList, q)
			}
		}
	}
	response := NewListQueueResponse(s.Config, filteredList)
	c.XML(200, response)
}

func (s *Server) removePermissions(c *gin.Context) {
	log.Println("RemovePermission is not implemented - Noop")
	c.XML(200, NewRemovePermissionResponse())
}

func (s *Server) sendMessage(c *gin.Context) {
	// DelaySeconds
	// MessageAttribute (Map)
	// MessageDeduplicationId
	// MessageGroupId
	// MessageBody TODOZ CHeck valid bytes
	// QueueUrl -> required
	body := c.Query("MessageBody")
	if len(body) == 0 {
		error := NewErrorResponse("Sender", "Invalid request: MissingQueryParamRejection(MessageBody)")
		c.XML(http.StatusBadRequest, error)
		return
	}
	url := c.Query("QueueUrl")
	if len(url) == 0 {
		error := NewErrorResponse("Sender", "Invalid request: MissingQueryParamRejection(QueueUrl)")
		c.XML(http.StatusBadRequest, error)
		return
	}
	messageID, err := s.Store.SendMessage(url, body)
	if err != nil {
		c.XML(http.StatusInternalServerError, NewErrorResponse("Sender", err.Error()))
		return
	}
	bodyMd5 := md5.Sum([]byte(body))
	// attrMd5 := md5.Sum([]byte(body))
	response := SendMessageResponse{
		SendMessageResult: SendMessageResult{
			MD5OfMessageBody:       fmt.Sprintf("%x", bodyMd5),
			MD5OfMessageAttributes: fmt.Sprintf("%x", "TODO"),
			MessageID:              messageID,
		},
		ResponseMetadata: ResponseMetadata{
			RequestId: DummyRequestID,
		},
	}
	c.XML(200, response)
}

func (s *Server) sendBadRequest(c *gin.Context) {
	c.String(http.StatusBadRequest, fmt.Sprintf("Unsupported request %s", c.Request.URL.String()))
}

func (s *Server) validateQueuName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("Queue name cannot be empty")
	}
	if len(name) > 80 {
		return fmt.Errorf("Queue name cannot be longer than 80 chars")
	}
	for _, c := range name {
		if !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '-' && c != '_' {
			return fmt.Errorf("Queue name may only contain letters, numbers, -, _")
		}
	}
	return nil
}
