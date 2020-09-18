import Util from "@/utils";

/****************************************************************
 * Define models for APIs
 ****************************************************************/

export class HostConfig {
  host: string = ""; // 호스트 도메인
  weight: number = 0; // 가중치 (정수)
}

export class HealthCheckConfig {
  url: string = ""; // 상태 검증용 URL
  timeout: any = "1s"; // 처리 제한 시간
  public AdjustSendValues() {
    this.timeout = Util.timeParser.ToDuration(this.timeout);
  }
  public AdjustReceiveValues() {
    this.timeout = Util.timeParser.FromDuration(this.timeout, "s");
  }
}

export class BackendConfig {
  hosts: Array<HostConfig> = [new HostConfig()]; // 백엔드 서비스 도메인 리스트 (Defintion에서 전역으로 설정한 경우는 생략 가능)
  timeout: any = "3s"; // 처리 제한 시간 (Defintion에서 전역으로 설정한 경우는 생략 가능)
  method: string = "GET"; // 호출 메서드 (Definition에서 전역으로 설정한 경우는 생략 가능)
  url_pattern: string = ""; // 서비스 URL (Domain 제외)
  encoding: string = "json"; // 데이터 인코딩 (xml, json : Definition에서 전역으로 설정한 경우는 생략 가능)
  group: string = ""; // 결과를 묶어서 처리할 그룹 명
  blacklist: Array<string> = []; // 결과에서 제외할 필드명 리스트 (flatmap 적용방식 "." 오퍼레이션 가능)
  whitelist: Array<string> = []; // 결과에서 포함할 필드명 리스트 (flatmap 적용방식 "." 오퍼레이션 가능)
  mapping: Map<string, string> = new Map(); // 결과에서 필드명을 변경할 리스트
  is_collection: boolean = false; // 결과가 Collection인지 여부
  wrap_collection_to_json: boolean = false; // 결과 Collection을 JSON의 "collection = []" 방식으로 처리할지 여부, false명 Array 상태로 처리
  target: string = ""; // 결과 중 특정한 필드만 처리할 경우의 필드명
  middleware?: any = {};
  disable_host_sanitize: boolean = false; // host 정보를 정제 작업할지 여부
  lb_mode: string = ""; // 백엔드 로드밸런싱 모드 ("rr", "wrr")

  public AdjustSendValues() {
    this.timeout = Util.timeParser.ToDuration(this.timeout);
  }
  public AdjustReceiveValues() {
    this.timeout = Util.timeParser.FromDuration(this.timeout, "s");
  }
}

export class ApiDefinition {
  name: string = ""; // 엔드포인트 식별 명
  active: boolean = false; // 엔드포인트 활성화 여부
  endpoint: string = ""; // 엔드포인트 URL
  hosts: Array<HostConfig> = [new HostConfig()]; // 백엔드 서비스 도메인 리스트
  method: string = "GET"; // 호출 메서드
  timeout: any = "1m"; // 처리 제한 시간
  cache_ttl: any = "3600s"; // 캐시 TTL
  output_encoding: string = "json"; // 데이터 인코딩 (xml, json)
  except_querystrings: Array<string> = []; // 벡엔드로 전달하지 않을 Query String 리스트
  except_headers: Array<string> = []; // 벡엔드로 전달하지 않을 Header 리스트
  middleware?: any = {};
  // health_check?: HealthCheckConfig = undefined; // 헬스 검증용 설정
  backend: Array<BackendConfig> = [new BackendConfig()]; // 백엔드 설정

  public AdjustSendValues() {
    this.timeout = Util.timeParser.ToDuration(this.timeout);
    this.cache_ttl = Util.timeParser.ToDuration(this.cache_ttl);
    this.backend.forEach(b => b.AdjustSendValues());
  }
  public AdjustReceiveValues() {
    this.timeout = Util.timeParser.FromDuration(this.timeout, "s");
    this.cache_ttl = Util.timeParser.FromDuration(this.cache_ttl);
    this.backend.forEach(b => b.AdjustReceiveValues());
  }
}

export class ApiGroup {
  name: string = "";
  definitions: Array<ApiDefinition> = [];

  public AdjustSendValues() {
    this.definitions.forEach(d => d.AdjustSendValues());
  }
  public AdjustReceiveValues() {
    this.definitions.forEach(d => d.AdjustReceiveValues());
  }
}

export function deserializeGroupFromJSON(
  val: any,
  isReceived: boolean = false
) {
  const group: ApiGroup = Object.assign(new ApiGroup(), val);
  group.definitions = group.definitions.map(d => {
    const def: ApiDefinition = Object.assign(new ApiDefinition(), d);
    def.backend = def.backend.map(b => Object.assign(new BackendConfig(), b));
    return def;
  });
  if (isReceived) group.AdjustReceiveValues();
  else group.AdjustSendValues();

  return group;
}

export function deserializeDefinitionFromJSON(
  val: any,
  isReceived: boolean = false
) {
  const def: ApiDefinition = Object.assign(new ApiDefinition(), val);
  def.backend = def.backend.map(b => Object.assign(new BackendConfig(), b));
  if (isReceived) def.AdjustReceiveValues();
  else def.AdjustSendValues();

  return def;
}
