# REST API G/W 테스트

> Notes
> ---
> **<font color="red">이 테스트 구성은 모두 Docker Container를 기준으로 하고 있으므로 사전에 docker 와 docker-compose 가 설치되어 있어야 합니다.</font>**

## 실행 방법

```shell
$ docker-compose up --build
```

상기의 명령으로 docker-compose 빌드 (--build) 진행 후에 실행 (up) 할 수 있습니다.
문제가 발생하면 바로 종료되므로 오류 메시지를 참고해서 문제를 해결하고 다시 실행하면 됩니다.

> Notes
> ---
> 변경된 내용이 없이 재 실행하는 경우는 Build 옵션을 사용하지 않아도 됩니다.
> ```shell
> $ docker-compose up
> ```
> 변경된 내용 (설정이나 소스 등)이 있는 경우는 반드시 Build 옵션을 사용해야 반영됩니다.

실행된 어플리케이션은 다음과 같습니다.

- **<font color="red">RESTAPIGW : localhost:8000</font>**
  - REST API G/W 서비스입니다. cb-spider 및 cb-tumblebug를 호출합니다.
- **<font color="red">InfluxDB : localhost:8086</font>**
  -  RESTAPIGW에서 Metrics 데이터를 저장하는 DB 서버입니다.
- **<font color="red">Grafana : localhost:3100</font>**
  - 수집된 Metrics 정보를 표시하는 UI 이므로 브라우저를 통해서 정보를 확인할 수 있습니다.
  - 초기 설정된 ID/PW 는 admin/admin 입니다.
  - 초기화면은 `Home Dashboard`입니다. 화면에 보이는 dashboard 리스트에서 `CB-RESTAPIGW`를 선택하시면 됩니다.
- **<font color="red">Jaeger : localhost:16686</font>**
  - RESTAPIGW에서 동작한 Trace 정보를 표시하는 수집기이며 UI를 제공하므로 브라우저를 통해서 정보를 확인할 수 있습니다.
  - 왼쪽의 `Search` 탭의 `Service` 부분에 cb-restapigw를 선택하고 `Find Traces` 버튼을 누르면 Trace 정보를 확인할 수 있습니다.
  - 단, Trace 수집 주기가 있으므로 초기에는 서비스가 보이지 않을 수 있습니다. refresh를 해서 서비스가 등록되었는지를 확인이 필요하며, 수집 주기 (10s) 이후에도 서비스가 등록되지 않았다면 터미널의 로그를 통해서 문제가 있는지를 확인해야 합니다.

## 테스트 방법

- POSTMAN Collection을 그대로 사용합니다.
- POSTMAN 환경설정의 ip, port를 위의 RESTAPIGW 에 맞도록 변경<font color="red">(localhost, 8000)</font>하시면 됩니다.
- 아래의 주의사항에 표시한 것과 같이 ./cb-restapigw/conf/cb-restapigw.yaml 파일에 cb-spider, cb-tumblebug Host Address를 (기본값은 CB-SPIDER : localhost:1024, CB-TUMBLEBUG : localhost:1323) 확인하시면 됩니다.

## 실행 중지

실행 상태인 터미널에서 `Ctrl+C` 로 중지 시그널을 처리하면 종료됩니다. 아래의 명령으로 사용된 리소스를 해제해 주시면 됩니다.
```shell
$ docker-compose down
```

> Notes
> ---
> 만일 터미널을 종료한 상태라면 다음과 같이 docker-compose.yaml 파일일 존재하는 폴더에서 터미널을 열고 아래의 명령을 실행하시면 됩니다.
> ```shell
> $ docker-compose stop   # docker-compose 종료
> $ docker-compose down   # 사용한 리소스 해제
> ```

## 주의 사항

- 만일 cb-spider 와 cb-tumblebug 의 address가 기본 설정 값과 다를 경우는 설정을 변경해 주셔야 합니다.
  - 설정 파일 경로 : ./cb-restapigw/conf/cb-restapigw.yaml
  - 변경 대상
    - cb-spider : http://localhost:1024 (기본 설정 값)
    - cb-tumblebug : http://localhost:1323 (기본 설정 값)
  - 변경된 설정 파일을 반영하려면 현재 버전에서는 다시 실행해야 합니다.

## 자체 연결 테스트 
- 모든 API 호출을 테스트한 것은 아니고 sampling으로 일부 API 호출 점검
- API G/W 에서 cb-spider 호출 (로그로 호출된 것 확인)
- API G/W 에서 cb-tumblebug 호출 (로그로 호출된 것 확인)
- cb-tumblebug에서 API G/W를 호출해서 cb-spider 호출 (tumblebug 호출은 로그로 확인되지만, 내부 오류로 인해서 진행 불가)
  - Nutsdb - not found bucket error
  - common.checkNS - not found bucket error
- cb-spider, cb-tumblebug에서 반환된 상태 코드가 200, 201이 아닌 경우는 API G/W 에서 500으로 처리 됨.

## 테스트 중에 발견된 문제점

자체 테스트 중에 발견된 문제점은 다음과 같습니다.

Gin Framework의 Route Handler 등록 시에 <font color="red">wildcard variable</font> 문제가 있습니다.

**이 문제는 HttpRoute 정책적인 (By Design) 문제로 다른 Framework (echo, mux 등)에도 동일하게 발생하는지를 검토한 후에 향후에는 다른 Route Engine으로 변경하는 것도 생각해 봐야할 것 같습니다.**

> 발생 대상
> ---
> - /ns/{ns_id}/mcis/<font color="red">recommend</font> 경로와
> - /ns/{ns_id}/mcis/<font color="red">{mcis_id}</font>/vm 경로의 모호성 때문에 발생합니다.

위의 예에서 "{mcis_id}" 는 path variable 로 실제 Path 구성 값으로 대체되는 변수입니다. 따라서 실제 path 값이 recommend 가 되면 두 개의 URL 이 동일한 상태가 되므로 Router 입장에서는 어떤 URL을 실행해야 할지 모르는 상황이 발생합니다. (뒤에 있는 추가 경로는 무시됩니다)

따라서 Gin Framework 에서는 path variable 을 기준으로 prefix를 비교하게 됩니다. 따라서 누가 먼저 등록되는지에 따라 그 다음에 등록되는 URL은 등록할 수 없는 문제가 발생합니다.
- recommend가 먼저 등록되면 {mcis_id}/vm 은 이미 해당 위치에 recommend (not path variable)가 사용되는 것으로 처리되어 있기 때문에 등록될 수 없습니다. (Tree Insert 관련 오류)
- {mcis_id}/vm 이 먼저 등록되면 recommend 위치에 {mcis_id}가 필요해서 등록할 수 없다는 것으로 처리됩니다. (Path Segment 관련 오류)

따라서 이번 테스트 설정에서는 이 문제를 회피하기 위해서 아래와 같이 URL을 변경했습니다.
<font color="red" weigth="bold">
CB-RESTAPIGW 설정의 Endpoint를 /ns/{ns_id}/mcis/recommend ==> /ns/{ns_id}/mcisrecommend로 설정하였습니다.</br>(cb-tumblebug 호출용 Backend url_pattern은 상관없음)</br>
따라서 테스트할 POSTMAN에서 이 호출 URL을 변경해 주시고 테스트 하시면 됩니다.
</font>