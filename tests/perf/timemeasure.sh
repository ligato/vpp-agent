export LC_NUMERIC="en_US.UTF-8"

calcTimeFormat(){
dt=$(echo "$1 - $2" | bc)
dd=$(echo "$dt/86400" | bc)
dt2=$(echo "$dt-86400*$dd" | bc)
dh=$(echo "$dt2/3600" | bc)
dt3=$(echo "$dt2-3600*$dh" | bc)
dm=$(echo "$dt3/60" | bc)
ds=$(echo "$dt3-60*$dm" | bc)
retstr=LC_NUMERIC="en_US.UTF-8" printf "%02d:%02d:%02.5f\n" $dh $dm $ds
echo $retstr
}

calcNanoTime(){
a=(`echo $1 | sed -e 's/[:]/ /g'`)
seconds= echo "${a[2]}+60*${a[1]}+3600*${a[0]}" | bc
echo $seconds
}

showTime(){
dd=$(echo "$1/86400" | bc)
dt2=$(echo "$1-86400*$dd" | bc)
dh=$(echo "$dt2/3600" | bc)
dt3=$(echo "$dt2-3600*$dh" | bc)
dm=$(echo "$dt3/60" | bc)
ds=$(echo "$dt3-60*$dm" | bc)
retstr=LC_NUMERIC="en_US.UTF-8" printf "%02d:%02d:%02.5f\n" $dh $dm $ds
echo $retstr
}

processResult(){
start=0
startAgent=0
resynctime=0
etcdTime=0
resyncTookTime=0
etcdTookTime=0
resyncStamp=0
etcdStamp=0
run=1
line_id=1
rel_item=0
echo "#record,#run,step,timeline,relative time,relative to #record,duration(ms)" > log/measuring_exp.csv
while IFS="," read -r val1 val2 val3;do
  if [ "$val2" == ' Started measuring' ]
  then
    start=$(calcNanoTime $val1)
    echo "$line_id,$run,Measuring started,$val1,00:00:00.0,0" >> log/measuring_exp.csv
    echo "$line_id,$run,Measuring started,$val1,00:00:00.0,0"
    rel_item=$line_id
    line_id=$((line_id+1))
  elif [ "$val2" == ' Starting the agent...' ]
  then
    startAgent=$(calcNanoTime $val1)
    diff=$(echo "$startAgent-$start" | bc)
    if [ $(bc <<< "$diff > 0") -eq 1 ]
    then
        time=$(showTime $diff)
        echo "$line_id,$run,Starting Agent,$val1,$time,$rel_item" >> log/measuring_exp.csv
        echo "$line_id,$run,Starting Agent,$val1,$time,$rel_item"
        rel_item=$line_id
        line_id=$((line_id+1))
    else
      time=$(showTime $start)
      echo "$line_id,$run,Kill failed,$time,00:00:00.0,0" >> log/measuring_exp.csv
      echo "$line_id,$run,Kill failed,$time,00:00:00.0,0"
      line_id=$((line_id+1))
    fi
  elif [[ "$val2" =~ ' Connecting to etcd took' ]]
  then
    etcdTime=$(calcNanoTime $val1)
    etcdStamp=$val1
    #etcdTookTime=$(bc <<< "scale = 10; $val3 / 1000000 ")
    a=(`echo $val3 | sed -e 's/[:]/ /g'`)
    etcdTookTime=$(bc <<< "scale = 10; ${a[0]} / 1000000 ")

  elif [[ "$val2" =~ ' Resync took' ]]
  then
    resyncTime=$(calcNanoTime $val1)
    resyncStamp=$val1
    #resyncTookTime=$(bc <<< "scale = 10; $val3 / 1000000 ")
    a=(`echo $val3 | sed -e 's/[:]/ /g'`)
    resyncTookTime=$(bc <<< "scale = 10; ${a[0]} / 1000000 ")

  elif [[ "$val2" =~ ' Connecting to VPP took' ]]
  then
    vppTime=$(calcNanoTime $val1)
    vppStamp=$val1
    #vppTookTime=$(bc <<< "scale = 10; $val3 / 1000000 ")
    a=(`echo $val3 | sed -e 's/[:]/ /g'`)
    vppTookTime=$(bc <<< "scale = 10; ${a[0]} / 1000000 ")
  elif [[ "$val2" =~ ' Connecting to kafka took' ]]
  then
    kafkaTime=$(calcNanoTime $val1)
    kafkaStamp=$val1
    #kafkaTookTime=$(bc <<< "scale = 10; $val3 / 1000000 ")
    a=(`echo $val3 | sed -e 's/[:]/ /g'`)
    kafkaTookTime=$(bc <<< "scale = 10; ${a[0]} / 1000000 ")
  elif [[ "$val2" =~ ' plugin GoVPP: Init' ]]
  then
    GoVPPInitTime=$(calcNanoTime $val1)
    GoVPPInitStamp=$val1
    a=(`echo $val3 | sed -e 's/[:]/ /g'`)
    GoVPPInitTookTime=$(bc <<< "scale = 10; ${a[0]} / 1000000 ")
  elif [[ "$val2" =~ ' plugin Linux: Init' ]]
  then
    LinuxInitTime=$(calcNanoTime $val1)
    LinuxInitStamp=$val1
    a=(`echo $val3 | sed -e 's/[:]/ /g'`)
    LinuxInitTookTime=$(bc <<< "scale = 10; ${a[0]} / 1000000 ")
  elif [[ "$val2" =~ ' plugin VPP: Init' ]]
  then
    PluginVPPInitTime=$(calcNanoTime $val1)
    PluginVPPInitStamp=$val1
    a=(`echo $val3 | sed -e 's/[:]/ /g'`)
    PluginVPPInitTookTime=$(bc <<< "scale = 10; ${a[0]} / 1000000 ")
  elif [[ "$val2" =~ ' resync the VPP Configuration end' ]]
  then
    PluginVPPResyncTime=$(calcNanoTime $val1)
    PluginVPPResyncStamp=$val1
    a=(`echo $val3 | sed -e 's/[:]/ /g'`)
    PluginVPPResyncTookTime=$(bc <<< "scale = 10; ${a[0]} / 1000000 ")
  elif [[ "$val2" =~ ' Agent Init' ]]
  then
    AgentInitTime=$(calcNanoTime $val1)
    AgentInitStamp=$val1
    a=(`echo $val3 | sed -e 's/[:]/ /g'`)
    AgentInitTookTime=$(bc <<< "scale = 10; ${a[0]} / 1000000 ")
  elif [[ "$val2" =~ ' Agent AfterInit' ]]
  then
    AgentAfterInitTime=$(calcNanoTime $val1)
    AgentAfterInitStamp=$val1
    a=(`echo $val3 | sed -e 's/[:]/ /g'`)
    AgentAfterInitTookTime=$(bc <<< "scale = 10; ${a[0]} / 1000000 ")
  elif [ "$val2" == ' Killed' ]
  then
    start1=$(calcNanoTime $val1)
    diff=$(echo "$start1-$start" | bc)
    if [ $(bc <<< "$diff < 0") -eq 1 ]
    then
      echo "$line_id,$run,Kill failed,$val1,$diff,$rel_item" >> log/measuring_exp.csv
      echo "$line_id,$run,Kill failed,$val1,$diff,$rel_item"
      line_id=$((line_id+1))
    fi
    run=$((run + 1))
    start=$start1
    echo "$line_id,$run,Container Killed,$val1,$diff,$rel_item" >> log/measuring_exp.csv
    echo "$line_id,$run,Container Killed,$val1,$diff,$rel_item"
    rel_item=$line_id
    line_id=$((line_id+1))
    startAgent=0
    resynctime=0
    etcdTime=0
    vppTime=0
    kafkaTime=0
    GoVPPInitTime=0
    LinuxInitTime=0
    PluginVPPInitTime=0
    PluginVPPResyncTime=0
    AgentInitTime=0
    AgentAfterInitTime=0
    resyncTookTime=0
    etcdTookTime=0
    vppTookTime=0
    kafkaTookTime=0
    AllInitTime=0
    GoVPPInitTookTime=0
    LinuxInitTookTime=0
    PluginVPPInitTookTime=0
    PluginVPPResyncTookTime=0
    AgentInitTookTime=0
    AgentAfterInitTookTime=0
    resyncStamp=0
    etcdStamp=0
    vppStamp=0
    kafkaStamp=0
    GoVPPInitStamp=0
    LinuxInitStamp=0
    PluginVPPInitStamp=0
    PluginVPPResyncStamp=0
    AgentInitStamp=0
    AgentAfterInitStamp=0
    AllInitStamp=0
  elif [ "$val2" == '--' ]
  then
    diff=$(echo "$startAgent-$start" | bc)
    if [ $(bc <<< "$diff > 0") -eq 1 ]
    then
      diff=$(echo "$etcdTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,ETCD connected,$etcdStamp,$time,$rel_item,$etcdTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,ETCD connected,$etcdStamp,$time,$rel_item,$etcdTookTime"
      line_id=$((line_id+1))
      diff=$(echo "$kafkaTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,Kafka connected,$kafkaStamp,$time,$rel_item,$kafkaTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,Kafka connected,$kafkaStamp,$time,$rel_item,$kafkaTookTime"
      line_id=$((line_id+1))
      diff=$(echo "$vppTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,VPP connected,$vppStamp,$time,$rel_item,$vppTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,VPP connected,$vppStamp,$time,$rel_item,$vppTookTime"
      line_id=$((line_id+1))
      diff=$(echo "$GoVPPInitTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,GoVPP Init,$GoVPPInitStamp,$time,$rel_item,$GoVPPInitTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,GoVPP Init,$GoVPPInitStamp,$time,$rel_item,$GoVPPInitTookTime"

      line_id=$((line_id+1))
      diff=$(echo "$LinuxInitTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,Linux plugin Init,$LinuxInitStamp,$time,$rel_item,$LinuxInitTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,Linux plugin Init,$LinuxInitStamp,$time,$rel_item,$LinuxInitTookTime"

      line_id=$((line_id+1))
      diff=$(echo "$PluginVPPInitTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,VPP plugin Init,$PluginVPPInitStamp,$time,$rel_item,$PluginVPPInitTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,VPP plugin Init,$PluginVPPInitStamp,$time,$rel_item,$PluginVPPInitTookTime"

      line_id=$((line_id+1))
      diff=$(echo "$PluginVPPResyncTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,Resync of VPP config,$PluginVPPResyncStamp,$time,$rel_item,$PluginVPPResyncTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,Resync of VPP config,$PluginVPPResyncStamp,$time,$rel_item,$PluginVPPResyncTookTime"



      line_id=$((line_id+1))
      diff=$(echo "$resyncTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,Resync done,$resyncStamp,$time,$rel_item,$resyncTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,Resync done,$resyncStamp,$time,$rel_item,$resyncTookTime"

      line_id=$((line_id+1))
      diff=$(echo "$AgentInitTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,Agent Init,$AgentInitStamp,$time,$rel_item,$AgentInitTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,Agent Init,$AgentInitStamp,$time,$rel_item,$AgentInitTookTime"

      line_id=$((line_id+1))
      diff=$(echo "$AgentAfterInitTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,Agent AfterInit,$AgentAfterInitStamp,$time,$rel_item,$AgentAfterInitTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,Agent AfterInit,$AgentAfterInitStamp,$time,$rel_item,$AgentAfterInitTookTime"


      line_id=$((line_id+1))
      diff=$(echo "$AllInitTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,Agent ready,$AllInitStamp,$time,$rel_item" >> log/measuring_exp.csv
      echo "$line_id,$run,Agent ready,$AllInitStamp,$time,$rel_item"
      line_id=$((line_id+1))
    fi
  elif [ "$val2" == ' All plugins initialized successfully' ]
  then
      AllInitTime=$(calcNanoTime $val1)
      AllInitStamp=$val1
      a=(`echo $val3 | sed -e 's/[:]/ /g'`)
  fi
done <$1
}

[ -z $BASH ] || shopt -s expand_aliases
alias BEGINCOMMENT="if [ ]; then"
alias ENDCOMMENT="fi"

#BEGINCOMMENT
rm -rf log 2>&1
mkdir log
sudo docker run -p 22379:2379 --name etcd -e ETCDCTL_API=3 -d \
    quay.io/coreos/etcd:v3.1.0 /usr/local/bin/etcd \
    -advertise-client-urls http://0.0.0.0:2379 \
    -listen-client-urls http://0.0.0.0:2379 > log/etcd.log 2>&1

#sudo docker run -p 22379:2379 --name etcd -d -e ETCDCTL_API=3 \
#        quay.io/coreos/etcd:v3.1.0 /usr/local/bin/etcd \
#        -advertise-client-urls http://0.0.0.0:2379 \
#        -listen-client-urls http://0.0.0.0:2379 > /dev/null
#   # dump etcd content to make sure that etcd is ready
#sleep 8s
sudo docker exec -it etcd etcdctl get --prefix ""
echo "Etcd started..."


sudo docker run -p 2181:2181 -p 9092:9092 --name kafka  -d\
 --env ADVERTISED_HOST=172.17.0.1 --env ADVERTISED_PORT=9092 spotify/kafka > log/kafka.log 2>&1

#sudo docker run -p 2181:2181 -p 9092:9092 --name kafka -d \
# --env ADVERTISED_HOST=0.0.0.0 --env ADVERTISED_PORT=9092 spotify/kafka > /dev/null
#    # list kafka topics to ensure that kafka is ready
sudo docker exec -it kafka /opt/kafka_2.11-0.10.1.0/bin/kafka-topics.sh --list --zookeeper localhost:2181 > /dev/null 2> /dev/null
echo "Kafka started..."


echo "Loading topology to ETCD..."
./topology-generate-routes.sh 1000

#BEGINCOMMENT
restime0=$(showTime $(date +%s.%N))
kubectl apply -f vnf-vpp.yaml
echo "$restime0, Started measuring" > log/out.csv
kubectl apply -f vswitch-vpp.yaml
echo "Collecting logs"
sleep 60s

#kubectl logs vswitch-vpp > vswitch-vpp.log
#grep -E 'Starting the agent...|Connecting to etcd took|error|Resync took|All plugins initialized successfully'  vswitch-vpp.log > res1.log
if [ -z "$1" ]
then
  cycle=1
elif [ $1 -gt 50 ]
then
  echo "Max cycle > 50, cycle set tp 50!"
  cycle=50
elif [ $1 -lt 1 ]
then
  cycle=50
else
  cycle=$1
fi

#for i in {1..3}
for (( i = 1; i <= $cycle; i++ ))
do
    kubectl logs vswitch-vpp > log/vswitch-vpp${i}.log
    grep -E 'Starting the agent...|Connecting to etcd took|Resync took|Connecting to VPP took|Connecting to kafka took|All plugins initialized successfully|plugin Linux: Init|plugin GoVPP: Init|plugin VPP: Init|resync the VPP Configuration end|Agent Init|Agent AfterInit'  log/vswitch-vpp${i}.log > log/log${i}.log
    #cat log${i}.log | sed -r 's/^time="[0-9-]{10} ([^"]+)".+ msg="([^"]+)".*$/\1, \2/' >> out.csv
    #cat log/log${i}.log | sed -r 's/^time="[0-9-]{10} ([^"]+)".+ msg="([^"]+)".*?(durationInNs=([0-9]+)\s$|\s$)/\1, \2, \4/' >> log/out.csv
    #cat log/log${i}.log | sed -r 's/^time="[0-9-]{10} ([^"]+)".+ msg="([^"]+)"[^=]+?(durationInNs=([0-9]+)[^=]*[ ]*$|\[ ]*$)/\1, \2, \4/' >> log/out.csv
    cat log/log${i}.log | sed -r 's/^time="[0-9-]{10} ([^"]+)".+ msg="([^"]+)"(([^"]*(durationInNs[:]?[ ]?=([0-9]+)))|(\s*))/\1, \2, \6/' >> log/out.csv
    echo "--,--,--,--------" >> log/out.csv
    #cat log/log${i}.log | sed -r 's/^time="[0-9-]{10} ([^"]+)".+ msg="([^"]+)".*?(timeInNs=([0-9]+)\s$|\s$)/\1, \2, \4/' >> log/out.csv
    sleep 1s
    echo "Killing the vswitch-vpp pod - run ${i}"
    restime0=$(showTime $(date +%s.%N))
    #echo "$restime0, Killed ${i}" >> out.csv
    echo "$restime0, Killed" >> log/out.csv
    kubectl exec vswitch-vpp kill 1
    #kubectl exec vswitch-vpp kill 1
    sleep 60s
done

cycle=$((cycle+1))
kubectl logs vswitch-vpp > log/vswitch-vpp${cycle}.log
grep -E 'Starting the agent...|Connecting to etcd took|Resync took|Connecting to VPP took|Connecting to kafka took|All plugins initialized successfully|plugin Linux: Init|plugin GoVPP: Init|plugin VPP: Init|resync the VPP Configuration end|Agent Init|Agent AfterInit'  log/vswitch-vpp${cycle}.log > log/log${cycle}.log
#cat log${cycle}.log | sed -r 's/^time="[0-9-]{10} ([^"]+)".+ msg="([^"]+)".*$/\1, \2/' >> out.csv
#cat log/log${cycle}.log | sed -r 's/^time="[0-9-]{10} ([^"]+)".+ msg="([^"]+)".*?(durationInNs=([0-9]+)\s$|\s$)/\1, \2, \4/' >> log/out.csv
#cat log/log${i}.log | sed -r 's/^time="[0-9-]{10} ([^"]+)".+ msg="([^"]+)"[^=]+?(durationInNs=([0-9]+)[^=]*[ ]*$|\[ ]*$)/\1, \2, \4/' >> log/out.csv
cat log/log${i}.log | sed -r 's/^time="[0-9-]{10} ([^"]+)".+ msg="([^"]+)"(([^"]*(durationInNs[:]?[ ]?=([0-9]+)))|(\s*))/\1, \2, \6/' >> log/out.csv
echo "--,--,--,---------" >> log/out.csv

#cat log/log${cycle}.log | sed -r 's/^time="[0-9-]{10} ([^"]+)".+ msg="([^"]+)".*?(timeInNs=([0-9]+)\s$|\s$)/\1, \2, \4/' >> log/out.csv
#ENDCOMMENT

processResult log/out.csv
#kubectl describe pods | grep 'Container ID:'
#ENDCOMMENT
./remove-pods.sh
zip -r logresult.zip log
