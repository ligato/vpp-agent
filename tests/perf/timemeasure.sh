#!/bin/bash

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
      echo "$line_id,$run,  ETCD connected,$etcdStamp,$time,$rel_item,$etcdTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,  ETCD connected,$etcdStamp,$time,$rel_item,$etcdTookTime"
      line_id=$((line_id+1))
      diff=$(echo "$kafkaTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,  Kafka connected,$kafkaStamp,$time,$rel_item,$kafkaTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,  Kafka connected,$kafkaStamp,$time,$rel_item,$kafkaTookTime"
      line_id=$((line_id+1))
      diff=$(echo "$vppTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,  VPP connected,$vppStamp,$time,$rel_item,$vppTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,  VPP connected,$vppStamp,$time,$rel_item,$vppTookTime"
      line_id=$((line_id+1))
      diff=$(echo "$GoVPPInitTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,  GoVPP Init,$GoVPPInitStamp,$time,$rel_item,$GoVPPInitTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,  GoVPP Init,$GoVPPInitStamp,$time,$rel_item,$GoVPPInitTookTime"

      line_id=$((line_id+1))
      diff=$(echo "$LinuxInitTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,  Linux plugin Init,$LinuxInitStamp,$time,$rel_item,$LinuxInitTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,  Linux plugin Init,$LinuxInitStamp,$time,$rel_item,$LinuxInitTookTime"

      line_id=$((line_id+1))
      diff=$(echo "$PluginVPPInitTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,  VPP plugin Init,$PluginVPPInitStamp,$time,$rel_item,$PluginVPPInitTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,  VPP plugin Init,$PluginVPPInitStamp,$time,$rel_item,$PluginVPPInitTookTime"

      line_id=$((line_id+1))
      diff=$(echo "$PluginVPPResyncTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,  Resync of VPP config,$PluginVPPResyncStamp,$time,$rel_item,$PluginVPPResyncTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,  Resync of VPP config,$PluginVPPResyncStamp,$time,$rel_item,$PluginVPPResyncTookTime"



      line_id=$((line_id+1))
      diff=$(echo "$resyncTime-$startAgent" | bc)
      time=$(showTime $diff)
      echo "$line_id,$run,  Resync done,$resyncStamp,$time,$rel_item,$resyncTookTime" >> log/measuring_exp.csv
      echo "$line_id,$run,  Resync done,$resyncStamp,$time,$rel_item,$resyncTookTime"

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


setup() {
    #cleanup prev results
    rm -f logresult.zip
    rm -rf log 2>&1
    mkdir log

    #start kafka + etcd and wait until they are ready
    sudo docker run -p 22379:2379 --name etcd -e ETCDCTL_API=3 -d \
        quay.io/coreos/etcd:v3.1.0 /usr/local/bin/etcd \
        -advertise-client-urls http://0.0.0.0:2379 \
        -listen-client-urls http://0.0.0.0:2379 > log/etcd.log 2>&1

    sudo docker exec -it etcd etcdctl get --prefix ""
    echo "Etcd started..."

    sudo docker run -p 2181:2181 -p 9092:9092 --name kafka  -d\
     --env ADVERTISED_HOST=172.17.0.1 --env ADVERTISED_PORT=9092 spotify/kafka > log/kafka.log 2>&1

    #    # list kafka topics to ensure that kafka is ready
    sudo docker exec -it kafka /opt/kafka_2.11-0.10.1.0/bin/kafka-topics.sh --list --zookeeper localhost:2181 > /dev/null 2> /dev/null
    echo "Kafka started..."


    echo "Loading topology to ETCD..."
    ./topology.sh
    #./topology-generate-routes.sh 0
    #./topology-generate-routes.sh 1000
    #./topology-generate-fib.sh 1000
}

enableCoreDumpInPod() {
cat <<EOF > /tmp/change_core_dump_path.sh
echo /tmp/cores/core.dump > /proc/sys/kernel/core_pattern
EOF
    mkdir /tmp/cores
    kubectl exec vswitch-vpp  -- chmod +x /tmp/change_core_dump_path.sh
    kubectl exec vswitch-vpp  -- bash -c /tmp/change_core_dump_path.sh

    kubectl exec vswitch-vpp  -- bash -c ulimit -c 100000
}
coreDumpVpp() {
    vpp_id_line=$(kubectl exec vswitch-vpp -- ps aux | grep /usr/bin/vpp)
    echo $vpp_id_line
    vpp_id=$(echo $vpp_id_line | awk '{print $2}')
    echo $vpp_id

    restime0=$(showTime $(date +%s.%N))
    echo "$restime0, VPP Killed" >> log/out.csv
    echo "Killing the vpp - run ${i}"
    kubectl exec vswitch-vpp -- kill -s ABRT $vpp_id
    last_vpp_core_time='00:00:00.00'
    for (( ig = 1; ig <= 280; ig++ ))
    do
      if [ -f  /tmp/cores/core.dump ]
      then
	vpp_core_final=$(stat -c %y /tmp/cores/core.dump)
	vpp_core_final_time=$(echo $vpp_core_final | awk '{print $2}')
	vpp_core_final_time_zone=$(echo $vpp_core_final | awk '{print $3}')
	tz_sign=${vpp_core_final_time_zone:0:1}
	tz_value=${vpp_core_final_time_zone:1:4}
	tz=$(echo "$tz_value/100" | bc)
	vpp_calc=$(calcNanoTime $vpp_core_final_time)
	tzsec=$(echo "$tz*3600" | bc)
	echo "Core_final:$vpp_core_final"
	echo "Core_final_time:$vpp_core_final_time"
	echo "Core_final_time_zone:$vpp_core_final_time_zone"
	echo "TZsign:$tz_sign"
	echo "TZValue:$tz_value"
	echo "TZ:$tz"
	echo "valc:$vpp_calc"
	echo "tzsec:$tzsec"
	if [ "$vpp_core_final_time" ==  "$last_vpp_core_time" ]
	then
          if [ "$tz_sign" ==  "+" ]
	  then
	    echo "substract"
	    timecalc=$(echo "$vpp_calc-$tzsec" | bc)
	    showtime=$(showTime $timecalc)
	  else
	    echo "adding"
	    timecalc=$(echo "$vpp_calc+$tzsec" | bc)
	    showtime=$(showTime $timecalc)
	  fi
	  echo "$showtime, Core dump done" >> log/out.csv
	  break
	else
	  last_vpp_core_time=$vpp_core_final_time
	fi
      else
	echo "Waiting for core dump file: $ig"
      fi
      sleep 0.1
    done

}


handleLogsFromOneRun() {
    kubectl logs vswitch-vpp > log/vswitch-vpp$1.log
    grep -E 'Starting the agent...|Connecting to etcd took|Resync took|Connecting to VPP took|Connecting to kafka took|All plugins initialized successfully|plugin Linux: Init|plugin GoVPP: Init|plugin VPP: Init|resync the VPP Configuration end|Agent Init|Agent AfterInit'  log/vswitch-vpp$1.log > log/log$1.log
    cat log/log$1.log | sed -r 's/^time="[0-9-]{10} ([^"]+)".+ msg="([^"]+)"(([^"]*(durationInNs[:]?[ ]?=([0-9]+)))|(\s*))/\1, \2, \6/' >> log/out.csv
    echo "--,--,--,--------" >> log/out.csv
}


#########################################
waitTime="30s"

if [ -z "$1" ]
then
  cycle=1
else
  cycle=$1
fi

setup

restime0=$(showTime $(date +%s.%N))
kubectl apply -f vnf-vpp.yaml
echo "$restime0, Started measuring" > log/out.csv
kubectl apply -f vswitch-vpp.yaml
echo "Collecting logs"
sleep ${waitTime}

enableCoreDumpInPod

for (( i = 1; i <= $cycle; i++ ))
do
    handleLogsFromOneRun ${i}
    sleep 1s

    coreDumpVpp

    echo "Killing the vswitch-vpp pod - run ${i}"
    restime0=$(showTime $(date +%s.%N))
    echo "$restime0, Killed" >> log/out.csv
    kubectl exec vswitch-vpp kill 1

    sleep ${waitTime}
done

handleLogsFromOneRun $((cycle+1))

processResult log/out.csv

zip -r logresult.zip log

./remove-pods.sh
