export namespace repeater {
	
	export class ParsedTarget {
	    scheme: string;
	    host: string;
	    port: number;
	
	    static createFrom(source: any = {}) {
	        return new ParsedTarget(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.scheme = source["scheme"];
	        this.host = source["host"];
	        this.port = source["port"];
	    }
	}
	export class ImportedRequest {
	    source: string;
	    requestText: string;
	    target: ParsedTarget;
	    hostHeaderMismatch: boolean;
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new ImportedRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = source["source"];
	        this.requestText = source["requestText"];
	        this.target = this.convertValues(source["target"], ParsedTarget);
	        this.hostHeaderMismatch = source["hostHeaderMismatch"];
	        this.warnings = source["warnings"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ParsedRequest {
	    method: string;
	    path: string;
	    version: string;
	    headers: string[][];
	    body: string;
	    target: ParsedTarget;
	    isBareUrl: boolean;
	    skeleton: string;
	
	    static createFrom(source: any = {}) {
	        return new ParsedRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.method = source["method"];
	        this.path = source["path"];
	        this.version = source["version"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.target = this.convertValues(source["target"], ParsedTarget);
	        this.isBareUrl = source["isBareUrl"];
	        this.skeleton = source["skeleton"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class RawResponse {
	    statusLine: string;
	    statusCode: number;
	    headersRaw: string;
	    bodyRaw: string;
	    bodyDecoded: string;
	    bytes: number;
	    durationMs: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new RawResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statusLine = source["statusLine"];
	        this.statusCode = source["statusCode"];
	        this.headersRaw = source["headersRaw"];
	        this.bodyRaw = source["bodyRaw"];
	        this.bodyDecoded = source["bodyDecoded"];
	        this.bytes = source["bytes"];
	        this.durationMs = source["durationMs"];
	        this.error = source["error"];
	    }
	}
	export class SendRequest {
	    requestText: string;
	    scheme: string;
	    host: string;
	    port: number;
	    useProxy: boolean;
	    proxyURL: string;
	    followHostHeader: boolean;
	    autoContentLength: boolean;
	    verifyTLS: boolean;
	    timeoutMs: number;
	    lang: string;
	
	    static createFrom(source: any = {}) {
	        return new SendRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestText = source["requestText"];
	        this.scheme = source["scheme"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.useProxy = source["useProxy"];
	        this.proxyURL = source["proxyURL"];
	        this.followHostHeader = source["followHostHeader"];
	        this.autoContentLength = source["autoContentLength"];
	        this.verifyTLS = source["verifyTLS"];
	        this.timeoutMs = source["timeoutMs"];
	        this.lang = source["lang"];
	    }
	}
	export class SendResult {
	    requestSent: string;
	    response?: RawResponse;
	
	    static createFrom(source: any = {}) {
	        return new SendResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestSent = source["requestSent"];
	        this.response = this.convertValues(source["response"], RawResponse);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

