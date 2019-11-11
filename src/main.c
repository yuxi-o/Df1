/*  Copyright (C) 2002 Stephane LEICHT (stephane.leicht@gmail.com)
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

#include "df1.h"

#define  DEFAULT_CONFIG_FILE	"/etc/df1d.conf"
extern word tns;
int file;
int Terminated=FALSE;
int df1_mode = 0; // 0: full-duplex, 1: half-duplex  Only support full-duplex now.

void Termine (int sig);

static int get_config(char *file, char *device, char *mode, int *speed, int *databits, int *parity, int *stopbits){
	int fd = open(file, O_RDONLY);	
	if(fd < 0){
		return(-1);
	}

	char config_line[64];
	read(fd, config_line, sizeof(config_line));
	int num = sscanf(config_line, "%s%s%d%d%d%d", device, mode, speed, databits, parity, stopbits);
	if(num != 6){
		return (-1);
	}

	close(fd);
	return 0;
}

//******************** MAIN ************
int main (int argc, char *argv[]) 
{
	openlog("DF1",LOG_NDELAY,LOG_DAEMON);
	setlogmask(~LOG_MASK(LOG_DEBUG)); // no debug informations

	MyLog("[v1.1.0] Usage: %s /dev/ttyxxx mode speed databits parity stopbits\n", argv[0]);
	MyLog("mode: specify the mode , full or half, not using, just for the same.\n");
	MyLog("speed: specify the bps, 115200, 57600, 9600, 4800, 2400...\n");
	MyLog("parity: 0:none, 1:even, 2:odd\n");
	MyLog("stopbits: 1 or 2\n");

	if(!((argc == 7) || (argc == 1))){
		MyLog( "Usage: %s /dev/ttyxxx mode speed databits parity stopbits\n", argv[0]);
		return(-1);
	}

	int speed = 0, databits = 0, parity = 0, stopbits = 0, num = 0;
	char device[32];
	char mode[16];

	if(argc == 7){
		speed = atoi(argv[3]);
		databits = argv[4][1] - '0';
		parity = argv[5][1] - '0';
		stopbits = argv[6][1] - '0';
		strncpy(device, argv[1], sizeof(device));
		strncpy(mode, argv[2], sizeof(mode));
	} else {
		if(0 != get_config(DEFAULT_CONFIG_FILE, device, mode, &speed, &databits, &parity, &stopbits))  
		// || (0 != get_config(argv[2], device, mode, &speed, &databits, &parity, &stopbits) ) 
		{
			MyLog("Config file doesn't exist or format error.\n");
			return(-1);
		}
	}
	
	signal(SIGTERM,Termine);
	signal(SIGINT,Termine);
	signal(SIGQUIT,Termine);
	signal(SIGSEGV,Termine);
	
//	if ((file=Df1_open_device ("/dev/ttyS0", 9600,0,8,1)) == -1)
	if ((file=Df1_open_device (device, speed, parity, databits, stopbits)) == -1)
	{
		MyLog("OpenCom Failed\n");
		return (-1);
	}

	if((strcasecmp("half", mode)) == 0){
		df1_mode = 1;
	}

#ifndef DEBUG	
	switch (fork())
	{
		case -1:
				syslog(LOG_NOTICE, "Error: creation df1 Daemon");
				//closelog;
				exit(1);
		case 0:
				setsid();
				chdir("/");
				umask(0);
				close(0);
				close(1);
				close(2);
				syslog(LOG_NOTICE,"Daemon OK");
				if (file == -1)
				{
					syslog(LOG_NOTICE,"OpenCom Failed\n");
					//closelog;
					exit(2);
				}
				else 
				{	
#endif
					server();	
					close (file);
					exit(0);
#ifndef DEBUG
				}
		default : exit(0);
	}
#endif	
}
//*********************************************
void Termine (int sig)
{
	switch (sig)
	{
		case SIGTERM:	syslog(LOG_NOTICE,"receive SIGTERM\n");
							Terminated=TRUE;
							break;
		case SIGINT:	syslog(LOG_NOTICE,"receive SIGINT\n");
							Terminated=TRUE;
							break;
		case SIGQUIT:	syslog(LOG_NOTICE,"receive SIGQUIT\n");
							Terminated=TRUE;
							break;
		case SIGSEGV:	syslog(LOG_NOTICE,"receive SIGSEGV, program ERROR\n");
							Terminated=TRUE;
							break;
		default:	syslog(LOG_NOTICE,"receive signal: %d\n",sig);
					Terminated=TRUE;
					break;
	}
}

