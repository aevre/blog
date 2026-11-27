package email

type Message struct {
	From    string
	To      []string
	Subject string
	HTML    string
	Text    string
}
