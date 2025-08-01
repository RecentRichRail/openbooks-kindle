package main

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/evan-buss/openbooks/server"
	"github.com/evan-buss/openbooks/util"

	"github.com/spf13/cobra"
)

var openBrowser = false
var serverConfig server.Config

func init() {
	desktopCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringVarP(&serverConfig.Port, "port", "p", "5228", "Set the local network port for browser mode.")
	serverCmd.Flags().IntP("rate-limit", "r", 10, "The number of seconds to wait between searches to reduce strain on IRC search servers. Minimum is 10 seconds.")
	serverCmd.Flags().BoolVar(&serverConfig.DisableBrowserDownloads, "no-browser-downloads", false, "The browser won't recieve and download eBook files, but they are still saved to the defined 'dir' path.")
	serverCmd.Flags().StringVar(&serverConfig.Basepath, "basepath", "/", `Base path where the application is accessible. For example "/openbooks/".`)
	serverCmd.Flags().BoolVarP(&openBrowser, "browser", "b", false, "Open the browser on server start.")
	serverCmd.Flags().BoolVar(&serverConfig.Persist, "persist", false, "Persist eBooks in 'dir'. Default is to delete after sending.")
	serverCmd.Flags().StringVarP(&serverConfig.DownloadDir, "dir", "d", filepath.Join(os.TempDir(), "openbooks"), "The directory where eBooks are saved when persist enabled.")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run OpenBooks in server mode.",
	Long:  "Run OpenBooks in server mode. This allows you to use a web interface to search and download eBooks.",
	PreRun: func(cmd *cobra.Command, args []string) {
		// Load environment variables from .env file if it exists
		log.Println("SERVER: About to load .env file...")
		
		// Try to load .env from the current directory first
		err := util.LoadEnvFile(".env")
		if err != nil {
			log.Printf("SERVER: Error loading .env file from current dir: %v", err)
		}
		
		// Also try to load from the executable's directory
		execPath, _ := os.Executable()
		execDir := filepath.Dir(execPath)
		envPath := filepath.Join(execDir, ".env")
		err = util.LoadEnvFile(envPath)
		if err != nil {
			log.Printf("SERVER: Error loading .env file from exec dir: %v", err)
		}
		
		log.Println("SERVER: .env file loading completed")
		
		bindGlobalServerFlags(&serverConfig)
		rateLimit, _ := cmd.Flags().GetInt("rate-limit")
		ensureValidRate(rateLimit, &serverConfig)
		
		// If cli flag isn't set (default value) check for the presence of an
		// environment variable and use it if found.
		if serverConfig.Basepath == cmd.Flag("basepath").DefValue {
			if envPath, present := os.LookupEnv("BASE_PATH"); present {
				serverConfig.Basepath = envPath
			}
		}
		serverConfig.Basepath = sanitizePath(serverConfig.Basepath)
		
		// Load SMTP configuration from environment variables
		serverConfig.SMTPHost = util.GetEnvString("SMTP_HOST", "")
		serverConfig.SMTPPort = util.GetEnvInt("SMTP_PORT", 587)
		serverConfig.SMTPUsername = util.GetEnvString("SMTP_USERNAME", "")
		serverConfig.SMTPPassword = util.GetEnvString("SMTP_PASSWORD", "")
		serverConfig.SMTPFrom = util.GetEnvString("SMTP_FROM", "")
		serverConfig.SMTPEnabled = util.GetEnvBool("SMTP_ENABLED", false)
		
		// Debug: Print SMTP configuration
		log.Printf("SMTP Configuration loaded:")
		log.Printf("  SMTP_ENABLED: %t", serverConfig.SMTPEnabled)
		log.Printf("  SMTP_HOST: %s", serverConfig.SMTPHost)
		log.Printf("  SMTP_PORT: %d", serverConfig.SMTPPort)
		log.Printf("  SMTP_USERNAME: %s", serverConfig.SMTPUsername)
		log.Printf("  SMTP_FROM: %s", serverConfig.SMTPFrom)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if openBrowser {
			browserUrl := "http://127.0.0.1:" + path.Join(serverConfig.Port+serverConfig.Basepath)
			util.OpenBrowser(browserUrl)
		}

		server.Start(serverConfig)
	},
}
