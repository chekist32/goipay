#!/bin/bash

protoPathDir="../proto"
protoGoOutDir="../internal/pb"
protoGoGrpcOutDir=$protoGoOutDir

while getopts ":i:o:" opt
do
  case "$opt" in
    i)
      protoPathDir="$OPTARG"
      ;;
    o)
      protoGoOutDir="$OPTARG"
      protoGoGrpcOutDir=$protoGoOutDir
      ;;
    ?)
      echo "Invalid option."
      exit 1
      ;;
  esac
done

shift $((OPTIND-1))

if [[ $# < 1 ]]
then
    echo "Too few args"
    exit 1
fi

while (( $# > 0 ))
do
    v=$1
    mkdir -p $protoGoOutDir/$v
    
    opts=""
    inputs=""
    for pb in $(find "$protoPathDir/$v" -name '*.proto')
    do
        pb=$(basename "$pb")
        opts+="--go-grpc_opt=M$pb=$protoGoOutDir/$v --go_opt=M$pb=$protoGoOutDir/$v "
        inputs+="$protoPathDir/$v/$pb "
    done

    protoc --proto_path="$protoPathDir/$v" \
        --go_out="$protoGoOutDir/$v" \
        --go-grpc_out="$protoGoGrpcOutDir/$v" \
        --go_opt="M$pb=$protoGoOutDir/$v" \
        --go_opt=paths=source_relative \
        --go-grpc_opt=paths=source_relative \
        $opts \
        $inputs
        
    shift
done