
# TODO

## Replicate relevant Prow jobs

### trusted infra jobs

```yaml
postsubmits:
  kubernetes/org:
  - name: post-org-peribolos
    cluster: test-infra-trusted
    decorate: true
    branches:
    - ^main$
    max_concurrency: 1
    spec:
      containers:
      - image: golang:1.17
        command:
        - ./admin/update.sh
        args:
        - --github-endpoint=http://ghproxy.default.svc.cluster.local
        - --github-endpoint=https://api.github.com
        - --github-token-path=/etc/github-token/oauth
        - --tokens=1200
        - --confirm
        volumeMounts:
        - name: github
          mountPath: /etc/github-token
      volumes:
      - name: github
        secret:
          secretName: oauth-token
    annotations:
      testgrid-alert-email: kubernetes-sig-testing-alerts@googlegroups.com, k8s-infra-oncall@google.com
      testgrid-num-failures-to-alert: '1'
...
periodics:
- interval: 24h
  name: ci-org-peribolos
  annotations:
    testgrid-dashboards: sig-contribex-org
    testgrid-tab-name: ci-peribolos
    testgrid-alert-email: kubernetes-github-managment-alerts@googlegroups.com, k8s-infra-oncall@google.com
    testgrid-num-failures-to-alert: '1'
  cluster: test-infra-trusted
  decorate: true
  max_concurrency: 1
  extra_refs:
  - org: kubernetes
    repo: org
    base_ref: main
  spec:
    containers:
    - image: golang:1.17
      command:
      - ./admin/update.sh
      args:
      - --github-endpoint=http://ghproxy.default.svc.cluster.local
      - --github-endpoint=https://api.github.com
      - --github-token-path=/etc/github-token/oauth
      - --tokens=1200
      - --confirm
      volumeMounts:
      - name: github
        mountPath: /etc/github-token
    volumes:
    - name: github
      secret:
        secretName: oauth-token
```

### k/org presubmits

```yaml
presubmits:
  kubernetes/org:
  - name: pull-org-test-all
    always_run: true
    decorate: true
    labels:
      preset-service-account: "true"
    spec:
      containers:
      - image: golang:1.17
        command:
        - make
        args:
        - test
    annotations:
      testgrid-num-columns-recent: '30'
      testgrid-create-test-group: 'true'
  - name: pull-org-verify-all
    always_run: true
    decorate: true
    labels:
      preset-service-account: "true"
    spec:
      containers:
      - image: golang:1.17
        command:
        - make
        args:
        - verify
    annotations:
      testgrid-num-columns-recent: '30'
      testgrid-create-test-group: 'true'
```
