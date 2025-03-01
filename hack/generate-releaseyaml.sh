#!/usr/bin/env bash
set -euf

export TARGET_REPO=${TARGET_REPO:-quay.io/openshift-pipeline/pipelines-as-code}
export TARGET_BRANCH=${TARGET_BRANCH:-main}
export TARGET_NAMESPACE=${TARGET_NAMESPACE:-pipelines-as-code}
export TARGET_OPENSHIFT=${TARGET_OPENSHIFT:-""}
export TARGET_PAC_VERSION=${PAC_VERSION:-"devel"}

MODE=${1:-""}

if [[ -n ${MODE} && ${MODE} == ko ]];then
    tmpfile=$(mktemp /tmp/.mm.XXXXXX)
    clean() { rm -f ${tmpfile}; }
    trap clean EXIT
    ko resolve -f config/ > ${tmpfile}

    if [[ ${TARGET_OPENSHIFT} != "" ]];then
       ko resolve -f config/openshift >> ${tmpfile}
    fi

    files="${tmpfile}"
else
    files=$(find config -maxdepth 1 -name '*.yaml'|sort -n)

    if [[ ${TARGET_OPENSHIFT} != "" ]];then
       files="${files} $(find config/openshift -maxdepth 1 -name '*.yaml'|sort -n)"
    fi
fi


for file in ${files};do
    head -1 ${file} | grep -q -- "---" || echo "---"
    sed -r -e "s,(.*image:.*)ko://github.com/openshift-pipelines/pipelines-as-code/cmd/pipelines-as-code-controller.*,\1${TARGET_REPO}:${TARGET_BRANCH}\"," \
        -e "s/(namespace: )\w+.*/\1${TARGET_NAMESPACE}/g" \
        -e "s,app.kubernetes.io/version:.*,app.kubernetes.io/version: \"${TARGET_PAC_VERSION}\"," \
        -e "/kind: Namespace$/ { n;n;s/name: .*/name: ${TARGET_NAMESPACE}/;}" \
        ${file}
done
