export namespace collectionsvc {
	
	export class Change {
	    key: string;
	    name: string;
	    set: string;
	    before: number;
	    after: number;
	
	    static createFrom(source: any = {}) {
	        return new Change(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.name = source["name"];
	        this.set = source["set"];
	        this.before = source["before"];
	        this.after = source["after"];
	    }
	}
	export class Diff {
	    total_before: number;
	    total_after: number;
	    changes: Change[];
	
	    static createFrom(source: any = {}) {
	        return new Diff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_before = source["total_before"];
	        this.total_after = source["total_after"];
	        this.changes = this.convertValues(source["changes"], Change);
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
	export class Query {
	    text: string;
	    set: string;
	    unresolved_only: boolean;
	    min_quantity: number;
	    sort: string;
	
	    static createFrom(source: any = {}) {
	        return new Query(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	        this.set = source["set"];
	        this.unresolved_only = source["unresolved_only"];
	        this.min_quantity = source["min_quantity"];
	        this.sort = source["sort"];
	    }
	}
	export class Row {
	    printing: string;
	    name: string;
	    set: string;
	    quantity: number;
	    unresolved: boolean;
	    raw?: string;
	    oracle?: string;
	    arena?: number;
	
	    static createFrom(source: any = {}) {
	        return new Row(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.printing = source["printing"];
	        this.name = source["name"];
	        this.set = source["set"];
	        this.quantity = source["quantity"];
	        this.unresolved = source["unresolved"];
	        this.raw = source["raw"];
	        this.oracle = source["oracle"];
	        this.arena = source["arena"];
	    }
	}

}

export namespace decks {
	
	export class OwnershipLine {
	    Name: string;
	    Required: number;
	    Owned: number;
	    Missing: number;
	    WildcardRelevant: number;
	
	    static createFrom(source: any = {}) {
	        return new OwnershipLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Required = source["Required"];
	        this.Owned = source["Owned"];
	        this.Missing = source["Missing"];
	        this.WildcardRelevant = source["WildcardRelevant"];
	    }
	}
	export class Ownership {
	    Snapshot: string;
	    Lines: OwnershipLine[];
	    TotalMissing: number;
	    Complete: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Ownership(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Snapshot = source["Snapshot"];
	        this.Lines = this.convertValues(source["Lines"], OwnershipLine);
	        this.TotalMissing = source["TotalMissing"];
	        this.Complete = source["Complete"];
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
	
	export class Problema {
	    Code: string;
	    Carta: string;
	    Detail: string;
	
	    static createFrom(source: any = {}) {
	        return new Problema(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Code = source["Code"];
	        this.Carta = source["Carta"];
	        this.Detail = source["Detail"];
	    }
	}
	export class RulesetID {
	    Format: string;
	    Version: number;
	
	    static createFrom(source: any = {}) {
	        return new RulesetID(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Format = source["Format"];
	        this.Version = source["Version"];
	    }
	}
	export class SubstitutionCandidate {
	    Name: string;
	    Score: number;
	    Evidence: string[];
	
	    static createFrom(source: any = {}) {
	        return new SubstitutionCandidate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Score = source["Score"];
	        this.Evidence = source["Evidence"];
	    }
	}
	export class Veredicto {
	    Ruleset: RulesetID;
	    Problemas: Problema[];
	    CatalogoStale: boolean;
	    Legal: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Veredicto(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Ruleset = this.convertValues(source["Ruleset"], RulesetID);
	        this.Problemas = this.convertValues(source["Problemas"], Problema);
	        this.CatalogoStale = source["CatalogoStale"];
	        this.Legal = source["Legal"];
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

export namespace homesvc {
	
	export class Action {
	    id: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new Action(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	    }
	}
	export class Model {
	    mtga_running: boolean;
	    source: string;
	    source_healthy: boolean;
	    last_sync?: string;
	    totals: number;
	    delta?: number;
	    state: viewstate.State;
	    actions: Action[];
	
	    static createFrom(source: any = {}) {
	        return new Model(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mtga_running = source["mtga_running"];
	        this.source = source["source"];
	        this.source_healthy = source["source_healthy"];
	        this.last_sync = source["last_sync"];
	        this.totals = source["totals"];
	        this.delta = source["delta"];
	        this.state = this.convertValues(source["state"], viewstate.State);
	        this.actions = this.convertValues(source["actions"], Action);
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

export namespace main {
	
	export class CollectionDiff {
	    diff: collectionsvc.Diff;
	    state: viewstate.State;
	
	    static createFrom(source: any = {}) {
	        return new CollectionDiff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.diff = this.convertValues(source["diff"], collectionsvc.Diff);
	        this.state = this.convertValues(source["state"], viewstate.State);
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
	export class CollectionPage {
	    rows: collectionsvc.Row[];
	    state: viewstate.State;
	
	    static createFrom(source: any = {}) {
	        return new CollectionPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rows = this.convertValues(source["rows"], collectionsvc.Row);
	        this.state = this.convertValues(source["state"], viewstate.State);
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
	export class DeckAnalysis {
	    preview: string;
	    verdict: decks.Veredicto;
	    ownership: decks.Ownership;
	    state: viewstate.State;
	
	    static createFrom(source: any = {}) {
	        return new DeckAnalysis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.preview = source["preview"];
	        this.verdict = this.convertValues(source["verdict"], decks.Veredicto);
	        this.ownership = this.convertValues(source["ownership"], decks.Ownership);
	        this.state = this.convertValues(source["state"], viewstate.State);
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
	export class DeckRevision {
	    id: string;
	    saved_at: string;
	    ruleset: string;
	
	    static createFrom(source: any = {}) {
	        return new DeckRevision(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.saved_at = source["saved_at"];
	        this.ruleset = source["ruleset"];
	    }
	}
	export class Substitution {
	    missing: string;
	    candidates: decks.SubstitutionCandidate[];
	
	    static createFrom(source: any = {}) {
	        return new Substitution(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.missing = source["missing"];
	        this.candidates = this.convertValues(source["candidates"], decks.SubstitutionCandidate);
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
	export class SubstitutionReport {
	    substitutions: Substitution[];
	    state: viewstate.State;
	
	    static createFrom(source: any = {}) {
	        return new SubstitutionReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.substitutions = this.convertValues(source["substitutions"], Substitution);
	        this.state = this.convertValues(source["state"], viewstate.State);
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

export namespace viewstate {
	
	export class State {
	    status: string;
	    code?: string;
	    detail?: string;
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.code = source["code"];
	        this.detail = source["detail"];
	    }
	}

}

