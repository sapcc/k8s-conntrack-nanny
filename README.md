k8s-conntrack-nanny
===================
This repository contains k8s-controller that purges conntrack entries for vanished UDP endpoints.
The controller watches for endpoint changes and whenever a UDP endpoint is removed from the list of active endpoints.

It is supposed to run as a Daemonset with `hostNetwork:true` and required `CAP_NET_ADMIN` privileges.

For an example chart see: https://github.com/sapcc/helm-charts/tree/master/system/kube-system/vendor/conntrack-nanny

Its a workaround for clusters <1.14 that are affected by this bug
https://github.com/kubernetes/kubernetes/issues/65228



