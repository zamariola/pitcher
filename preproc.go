package pitcher

import "log/slog"

func JWTAuth(req *Request, session Session) error {

	jwt, ok := session.Get(jwtKey)

	if !ok {
		slog.Warn("unable to get jwt token from session, skipping it")
		return nil
	}

	req.Headers.Add(authorizationKey, "Bearer "+jwt)
	return nil
}
