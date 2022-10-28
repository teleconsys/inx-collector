package api

func (s *Server) apiLogStart(route string) {
	s.WrappedLogger.LogInfof("Serving RestAPI '%s' request ... ", route)
}

func (s *Server) apiLogEnd(route string, err error) {
	if err != nil {
		s.WrappedLogger.LogErrorf("Serving RestAPI '%s' request ... failed, error: %w", route, err)
	} else {
		s.WrappedLogger.LogInfof("Serving RestAPI '%s' request ... done", route)
	}
}
