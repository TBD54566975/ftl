package namedext

// EmailConsent indicates wether someone has opted in or out of email communications.
// Consent is important and this is a second line of comments.
//
//ftl:enum
type EmailConsent string

const (
	EmailConsentOptIn  EmailConsent = "opt_in"
	EmailConsentOptOut EmailConsent = "opt_out"
)

// Comment typealias has a comment explaining what it is
//
//ftl:typealias
type Comment string

// Shared should export consent and comment
//
//ftl:data export
type Shared struct {
	Consent EmailConsent
	Comment Comment
}
