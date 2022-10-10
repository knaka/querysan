export namespace qsfts {
	
	export class QueryResult {
	    path: string;
	    title: string;
	    offsets: string;
	    snippet: string;
	
	    static createFrom(source: any = {}) {
	        return new QueryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.title = source["title"];
	        this.offsets = source["offsets"];
	        this.snippet = source["snippet"];
	    }
	}

}

