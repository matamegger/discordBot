package main

func generateString(strings ...string) (output string) {
	if len(strings) < 1 {
		return
	}
	output = strings[0]
	for i := 1; i < len(strings); i++ {
		output += " " + strings[i]
	}
	return
}
