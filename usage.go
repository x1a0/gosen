package main

func Usage(cmd string) string {
	switch cmd {
	case "gosen":
		return `
Usage:
    gosen command [arguments]

Available commands:
    login     login your Sony Entertainment Network
    friend    add friend
  `

	case "friend":
		return `
Usage:
    gosen friend [-m "optional message"] name1[, name2, ...]
  `

	default:
		return ""
	}
}
