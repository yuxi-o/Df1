#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <errno.h>
#include <ctype.h>

//#define SERV_IP "127.0.0.1"
#define SERV_PORT 18111
#define BSIZE	1024 

char serv_ip[32]={0};
int main(int argc, char* argv[])
{
	int fd;
	int ret = 0;
	struct sockaddr_in saddr;
	char buf[BSIZE];
	char filetype[] = "OISBTCRNF";

	if(!((argc == 3) || (argc == 4))){
		fprintf(stdout, "[v1.1.0] Usage: %s server_ip addr[=value] [salve_addr]\n", argv[0]);
		exit(-6);
	}

	if(strchr(filetype, toupper(argv[2][0])) == NULL){
		fprintf(stdout, "PLC addr format error! The addr begins with one letter of (OISBTCRNF).\n");
		exit(-6);
	}

	fd = socket(AF_INET, SOCK_STREAM, 0);
	if (fd == -1){
		fprintf(stdout, "socket error\n");
		exit(EXIT_FAILURE);
	}
	
	strcpy(serv_ip, argv[1]);
	memset(&saddr, 0, sizeof(struct sockaddr_in));
	saddr.sin_family = AF_INET;
	inet_pton(AF_INET, serv_ip, &saddr.sin_addr.s_addr);
	saddr.sin_port = htons(SERV_PORT);
	ret = connect(fd, (struct sockaddr *)&saddr, sizeof(struct sockaddr));
	if (fd == -1){
		fprintf(stdout, "connect error\n");
		exit(-20);
	}

//	while(1)
	{
		ret = write(fd, argv[2], strlen(argv[2])+1);
		if(ret == -1){
			if(errno == EPIPE){
				close(fd);
				fprintf(stdout, "[%s:%d] disconnect\n", serv_ip, SERV_PORT);
				exit(-20);
			}
		}
		memset(buf, 0, sizeof(buf));
		ret = read(fd, buf, sizeof(buf));
		if (ret > 0){
			int slen = strlen("error:");
			if(strncmp("error:", buf, slen) == 0){
				fprintf(stdout, "receive error from daemon\n");
				exit(atoi(&buf[slen]));
			}
			fprintf(stdout, "%s", buf);
			fflush(stdout);
		} else if(ret == 0){
			fprintf(stdout, "disconnect\n");
			exit(-20);
			// break;
		} else {
		/*	if (errno == EINTR){
				continue;
			}
			*/
			fprintf(stdout, "Unknown error\n");
			exit(-20);
		}

	}

	close(fd);
	return 0;
}
