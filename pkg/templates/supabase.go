package templates

import (
	"encoding/json"
	"fmt"

	"github.com/sheyaln/sabokit-broadside/pkg/broadside_mjml"
)

// parseEmailTreeJSON parses an email tree JSON string into an EmailBlock
func parseEmailTreeJSON(jsonStr string) (broadside_mjml.EmailBlock, error) {
	var rawData map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract the emailTree field
	emailTreeData, exists := rawData["emailTree"]
	if !exists {
		// If no emailTree wrapper, assume the entire JSON is the tree
		emailTreeData = []byte(jsonStr)
	}

	// Use the UnmarshalEmailBlock function from the broadside_mjml package
	emailBlock, err := broadside_mjml.UnmarshalEmailBlock(emailTreeData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal email block: %w", err)
	}

	return emailBlock, nil
}

// AllSupabaseTemplates returns a map of all Supabase template creation functions
// This makes it easy to iterate over all templates in tests
func AllSupabaseTemplates() map[string]func() (broadside_mjml.EmailBlock, error) {
	return map[string]func() (broadside_mjml.EmailBlock, error){
		"signup":           CreateSupabaseSignupEmailStructure,
		"magic_link":       CreateSupabaseMagicLinkEmailStructure,
		"recovery":         CreateSupabaseRecoveryEmailStructure,
		"email_change":     CreateSupabaseEmailChangeEmailStructure,
		"invite":           CreateSupabaseInviteEmailStructure,
		"reauthentication": CreateSupabaseReauthenticationEmailStructure,
	}
}
