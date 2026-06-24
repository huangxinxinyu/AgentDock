package config

import "os"

func FromOS() map[string]string {
	keys := []string{
		envHTTPAddr,
		envDatabaseURL,
		envAppSecret,
		envEncryptionKey,
		envCORSAllowedOrigin,
		envSandboxProvider,
		envAgentOSImage,
		envAgentOSWorkdir,
		envDockerNetwork,
		envDockerVolumePref,
	}
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		values[key] = os.Getenv(key)
	}
	return values
}
