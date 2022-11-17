package main

var dropHttpRespBodyPaths = map[string]bool{
	"/api/v0/cat": true,
}

func shouldDropHttpRespBody(path string) bool {
	if _, ok := dropHttpRespBodyPaths[path]; ok {
		return true
	}
	return false
}
