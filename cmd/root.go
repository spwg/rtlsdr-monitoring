/*
Copyright © 2024 Spencer Greene

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	wantDeviceFlagName = "device_name"
	frequencyFlagName  = "frequency"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rtlsdrmonitoring",
	Short: "Monitors software defined radio",
	Long: `rtlsdrmonitoring is intended to monitor software defined radios plugged into USB ports on a raspberry pi.
Running this in a systemd unit enables sending email notifications when something goes wrong with the device.`,
}

var watchCmd = &cobra.Command{
	Use:     "watchusbdevices --device_name=<device name>",
	Short:   "Watches usb devices to make sure a device is listed.",
	Example: "watchusbdevices --device_name='Realtek Semiconductor Corp. RTL2838 DVB-T'",
	Run: func(cmd *cobra.Command, args []string) {
		want, err := cmd.Flags().GetString(wantDeviceFlagName)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			os.Exit(1)
		}
		frequency, err := cmd.Flags().GetDuration(frequencyFlagName)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			os.Exit(1)
		}
		// Create a channel to control when the check happens.
		var ch <-chan time.Time
		if frequency > 0 {
			ticker := time.NewTicker(frequency)
			defer ticker.Stop()
			ch = ticker.C
		} else {
			// In oneshot mode, close the channel so the first read will return false.
			c := make(chan time.Time)
			close(c)
			ch = c
		}
		for {
			lsusb := exec.Command("lsusb")
			output, err := lsusb.Output()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "%q error: %v\n", lsusb.String(), err)
				os.Exit(1)
			}
			if !strings.Contains(string(output), want) {
				fmt.Fprintf(cmd.ErrOrStderr(), "%q output:\n%s\n", lsusb.String(), output)
				os.Exit(1)
			}
			_, ok := <-ch
			if !ok {
				break
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	watchCmd.Flags().String(wantDeviceFlagName, "", "Device name to ensure is found.")
	_ = watchCmd.MarkFlagRequired(wantDeviceFlagName)
	watchCmd.Flags().Duration(frequencyFlagName, 0, "How often to check. If 0 (default), just runs once.")
	rootCmd.AddCommand(watchCmd)
}
