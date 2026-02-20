package api

import ( 
	"regexp" 
	"strings" 
) 

var blockedDomains = map[string]bool{ 
	"mailinator.com": true, 
	"tempmail.com": true, 
	"10minutemail.com": true, 
	"guerrillamail.com": true, 
	"yopmail.com": true, 
	"trashmail.com": true, 
	"fakeinbox.com": true, 
	"dispostable.com": true, 
} 

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`) 

func IsValidEmailFormat(email string) bool { 
	return emailRegex.MatchString(email) 
} 

func IsDisposableEmail(email string) bool { 
	parts := strings.Split(email, "@") 
	if len(parts) != 2 { 
		return true 
	} 
	
	domain := strings.ToLower(parts[1]) 
	return blockedDomains[domain] 
}