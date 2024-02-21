./build4pi.sh

ssh pimyhome 'systemctl stop sync-dr-schedule'
sleep 3
scp dr-schedule-pi pimyhome:/home/osmc/sync-dr-schedule
ssh pimyhome 'systemctl start sync-dr-schedule'
