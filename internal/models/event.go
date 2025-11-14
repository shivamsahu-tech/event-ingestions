package models

type Event struct {
    SiteID    string `json:"site_id"`
    EventType string `json:"event_type"`
    Path      string `json:"path"`
    UserID    string `json:"user_id"`
    Timestamp string `json:"timestamp"`
}
