package nutrition

import "time"

// SendMessageRequest: POST /api/nutrition/messages
// Day is the user's LOCAL weekday ("monday"...), so "today's workout" is right for their timezone.
type SendMessageRequest struct {
	Content string `json:"content" validate:"required,notblank,max=1000"`
	Day     string `json:"day"     validate:"omitempty,oneof=monday tuesday wednesday thursday friday saturday sunday"`
}

type MessageResponse struct {
	ID        string    `json:"id"`
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewMessageResponse(m *Message) MessageResponse {
	return MessageResponse{ID: m.ID.Hex(), Role: m.Role, Content: m.Content, CreatedAt: m.CreatedAt}
}

type ListResponse struct {
	Items []MessageResponse `json:"items"`
}

type SendResponse struct {
	UserMessage      MessageResponse `json:"userMessage"`
	AssistantMessage MessageResponse `json:"assistantMessage"`
}
