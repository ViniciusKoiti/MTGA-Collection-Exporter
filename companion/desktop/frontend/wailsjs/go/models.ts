export namespace main {
	
	export class View {
	    id: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new View(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	    }
	}

}

export namespace setup {
	
	export class PrivacyChoices {
	    telemetry_opt_in: boolean;
	    assistant_enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PrivacyChoices(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.telemetry_opt_in = source["telemetry_opt_in"];
	        this.assistant_enabled = source["assistant_enabled"];
	    }
	}
	export class Result {
	    source: string;
	    needs_guidance: boolean;
	    verified: boolean;
	    fell_back: boolean;
	    privacy: PrivacyChoices;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = source["source"];
	        this.needs_guidance = source["needs_guidance"];
	        this.verified = source["verified"];
	        this.fell_back = source["fell_back"];
	        this.privacy = this.convertValues(source["privacy"], PrivacyChoices);
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

