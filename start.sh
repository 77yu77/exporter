grobalIP=$1
id=$2
ifconfig eth0:sdneth0 $grobalIP netmask 255.255.255.255 up
iptables -t nat -A OUTPUT -d 10.233.0.0/16 -j MARK --set-mark $id
iptables -t nat -A POSTROUTING -m mark --mark $id -d 10.233.0.0/16 -j SNAT --to-source $grobalIP
chmod 777 link_exporter
./link_exporter &
/podserver