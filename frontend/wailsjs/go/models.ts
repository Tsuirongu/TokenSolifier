export namespace app {
	
	export class PluginCreationRequest {
	    name: string;
	    chineseName: string;
	    description: string;
	    userRequirement: string;
	    tagIds: number[];
	
	    static createFrom(source: any = {}) {
	        return new PluginCreationRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.chineseName = source["chineseName"];
	        this.description = source["description"];
	        this.userRequirement = source["userRequirement"];
	        this.tagIds = source["tagIds"];
	    }
	}

}

export namespace models {
	
	export class ChatMessage {
	    role: string;
	    content: string;
	    // Go type: time
	    timestamp: any;
	
	    static createFrom(source: any = {}) {
	        return new ChatMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
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
	export class ChatSession {
	    id: string;
	    messages: ChatMessage[];
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ChatSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.messages = this.convertValues(source["messages"], ChatMessage);
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class ChatResponse {
	    sessionId: string;
	    message: ChatMessage;
	    session?: ChatSession;
	
	    static createFrom(source: any = {}) {
	        return new ChatResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.message = this.convertValues(source["message"], ChatMessage);
	        this.session = this.convertValues(source["session"], ChatSession);
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
	
	export class ClipboardItem {
	    id: number;
	    type: string;
	    content: string;
	    preview: string;
	    size: number;
	    is_fav: boolean;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new ClipboardItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.content = source["content"];
	        this.preview = source["preview"];
	        this.size = source["size"];
	        this.is_fav = source["is_fav"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	export class SelectOption {
	    label: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new SelectOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.value = source["value"];
	    }
	}
	export class InputItem {
	    content: string;
	    contentType: string;
	    description: string;
	    options?: SelectOption[];
	
	    static createFrom(source: any = {}) {
	        return new InputItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.content = source["content"];
	        this.contentType = source["contentType"];
	        this.description = source["description"];
	        this.options = this.convertValues(source["options"], SelectOption);
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
	export class Tag {
	    id: number;
	    name: string;
	    color: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Tag(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.color = source["color"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class PluginRelatedPlugin {
	    id: number;
	    pluginId: number;
	    relatedPluginId: number;
	    relatedPluginName: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new PluginRelatedPlugin(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.pluginId = source["pluginId"];
	        this.relatedPluginId = source["relatedPluginId"];
	        this.relatedPluginName = source["relatedPluginName"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class PluginDependency {
	    id: number;
	    pluginId: number;
	    package: string;
	    version: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new PluginDependency(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.pluginId = source["pluginId"];
	        this.package = source["package"];
	        this.version = source["version"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class PluginOutputDesc {
	    textList: InputItem[];
	    imageList: InputItem[];
	    fileList: InputItem[];
	    audioList: InputItem[];
	    videoList: InputItem[];
	    documentList: InputItem[];
	    otherList: InputItem[];
	    pluginError: string;
	
	    static createFrom(source: any = {}) {
	        return new PluginOutputDesc(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.textList = this.convertValues(source["textList"], InputItem);
	        this.imageList = this.convertValues(source["imageList"], InputItem);
	        this.fileList = this.convertValues(source["fileList"], InputItem);
	        this.audioList = this.convertValues(source["audioList"], InputItem);
	        this.videoList = this.convertValues(source["videoList"], InputItem);
	        this.documentList = this.convertValues(source["documentList"], InputItem);
	        this.otherList = this.convertValues(source["otherList"], InputItem);
	        this.pluginError = source["pluginError"];
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
	export class PluginInputDesc {
	    textList: InputItem[];
	    imageList: InputItem[];
	    fileList: InputItem[];
	    audioList: InputItem[];
	    videoList: InputItem[];
	    documentList: InputItem[];
	    otherList: InputItem[];
	
	    static createFrom(source: any = {}) {
	        return new PluginInputDesc(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.textList = this.convertValues(source["textList"], InputItem);
	        this.imageList = this.convertValues(source["imageList"], InputItem);
	        this.fileList = this.convertValues(source["fileList"], InputItem);
	        this.audioList = this.convertValues(source["audioList"], InputItem);
	        this.videoList = this.convertValues(source["videoList"], InputItem);
	        this.documentList = this.convertValues(source["documentList"], InputItem);
	        this.otherList = this.convertValues(source["otherList"], InputItem);
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
	export class Plugin {
	    id: number;
	    name: string;
	    chineseName: string;
	    description: string;
	    code: string;
	    isActive: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    input?: PluginInputDesc;
	    output?: PluginOutputDesc;
	    dependencies?: PluginDependency[];
	    relatedPlugins?: PluginRelatedPlugin[];
	    tags?: Tag[];
	
	    static createFrom(source: any = {}) {
	        return new Plugin(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.chineseName = source["chineseName"];
	        this.description = source["description"];
	        this.code = source["code"];
	        this.isActive = source["isActive"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.input = this.convertValues(source["input"], PluginInputDesc);
	        this.output = this.convertValues(source["output"], PluginOutputDesc);
	        this.dependencies = this.convertValues(source["dependencies"], PluginDependency);
	        this.relatedPlugins = this.convertValues(source["relatedPlugins"], PluginRelatedPlugin);
	        this.tags = this.convertValues(source["tags"], Tag);
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

export namespace services {
	
	export class PluginGenerationConfig {
	    UseAvailablePlugins: boolean;
	    UseAICapabilities: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PluginGenerationConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.UseAvailablePlugins = source["UseAvailablePlugins"];
	        this.UseAICapabilities = source["UseAICapabilities"];
	    }
	}

}

