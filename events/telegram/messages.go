package telegram

const msgHelp = `I'm your personal reading assistant.

Send me any URL – I'll keep it safe.
Use /rnd to get a random saved page.
Important: after that, it's gone from your list!`

const msgHello = "Hello! 👋\n\n" + msgHelp

const (
	msgUnknownCommand = "Command not found ❓"
	msgNoSavedPages   = "Nothing saved yet 📭"
	msgSaved          = "Saved! 💾"
	msgAlreadyExists  = "You've already got this one 📌"
)
