#!/usr/bin/env bash

# 遍历 xvideos.com 视频 ID，查找出对应用户的视频

users="chicken1806\|hushixiaolu"

id=75180633
ending=99999999

# do don't edit
header="User-Agent: Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Googlebot/2.1; +http://www.google.com/bot.html) Chrome/W.X.Y.Z Safari/537.36"

while true; do
  echo -e "\033[32m$id\033[0m"

  video=$(curl -k -H $header "https://www.xvideos.com/video$id/_" | grep 'setUploaderName\|setVideoURL' | sed "s/.*'\([^';']*\)'.*/\1/" | awk -v ORS=' ' '{print}')
  if [[ -n $(echo $video | grep "$users") ]]; then
    user=$(echo $video | awk '{print $1}')
    echo $video | awk '{print "https://www.xvideos.com"$2}' >> $user.txt
  fi

  start=$[$id+1]
done
