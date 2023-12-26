#!/bin/bash
## $1:passwd $2:docker image id
#4ece41bca6dc
nums=(4 5 7 8 9 0 12 13)
HOST=("10.0.0.14" "10.0.0.15" "10.0.0.17" "10.0.0.18" "10.0.0.19" "10.0.0.10" "10.0.0.22" "10.0.0.23")
for((i=0;i<${#nums[*]};i++));
do
USER=kube${nums[i]}
HO=${HOST[i]}
expect << EOF
spawn ssh $USER@$HO
expect "password:"
send "$1\r"
expect "$USER"
send "killall -9 monitor\r"
expect eof
EOF
done 

for((i=0;i<${#nums[*]};i++));
do
    USER=kube${nums[i]}
    lftp sftp://$USER:$1@${HOST[i]} <<EOF
    put monitor
    chmod 777 monitor
    bye
EOF
done 

for((i=0;i<${#nums[*]};i++));
do
USER=kube${nums[i]}
HO=${HOST[i]}
expect << EOF
spawn ssh $USER@$HO
expect "password:"
send "$1\r"
expect "$USER"
send "nohup ./monitor Network1 $2 &\r"
expect eof
EOF
done 
