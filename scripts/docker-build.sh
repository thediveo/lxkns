#!/bin/bash

dockerfile="$1"
args="${@:2}"

MAINUSE=./lxkns # <-- adapt this to reflect the "main" module
EMPTYCONTEXT=.emptyctx
NUMCONTEXTS=9

# find out if we are in workspace mode -- and it we are, then the list of
# modules actually used.
mkdir -p ${EMPTYCONTEXT}
trap 'rm -rf -- "${EMPTYCONTEXT}"' EXIT

contexts=()
workspace_details=$(go work edit --json)
if [[ ${workspace_details} ]]; then
    goworkdir=$(dirname $(go env GOWORK))
    echo "found workspace" ${goworkdir}
    diskpaths=$(echo ${workspace_details} | jq --raw-output '.Use | .[]? | .DiskPath')
    echo "modules used in workspace:" ${diskpaths}
    while IFS= read -r module; do
        if [[ "${module}" == "${MAINUSE}" ]]; then
            echo "  🏠" ${module};
        else
            relcontext=$(realpath --relative-to="." ${goworkdir}/${module})
            contexts+=( ${relcontext} )
            echo "  🧩" ${module} "» 📁" ${relcontext}
        fi
    done <<< ${diskpaths}
else
    diskpaths="${MAINUSE}"
fi

buildctxargs=()
buildargs=()
ctxno=1
for ctx in "${contexts[@]}"; do
    buildctxargs+=( "--build-context=bctx${ctxno}=${ctx}" )
    buildargs+=( "--build-arg=MOD${ctxno}=./$(basename ./${ctx})/" )
    ((ctxno=ctxno+1))
done
for ((;ctxno<=NUMCONTEXTS;ctxno++)); do
    buildctxargs+=( "--build-context=bctx${ctxno}=${EMPTYCONTEXT}" )
done
echo "args:" ${buildctxargs[*]} ${buildargs[*]}
echo "build inside:" ${CWD}

docker build \
    -f ${dockerfile} \
    ${buildargs[@]} \
    ${buildctxargs[@]} \
    --build-arg=WSDISKPATHS="$(echo ${diskpaths})" \
    ${args} \
    .
