export namespace cms {
	
	export class Abteilung {
	    ID: string;
	    Name: string;
	
	    static createFrom(source: any = {}) {
	        return new Abteilung(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	    }
	}
	export class Angebot {
	    ID: string;
	    Title: string;
	    Subtitle: sql.NullString;
	    // Go type: time
	    DateStart: any;
	    // Go type: time
	    DateStop: any;
	    Link: string;
	    Image: string;
	    Anzeigen: sql.NullBool;
	
	    static createFrom(source: any = {}) {
	        return new Angebot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Title = source["Title"];
	        this.Subtitle = this.convertValues(source["Subtitle"], sql.NullString);
	        this.DateStart = this.convertValues(source["DateStart"], null);
	        this.DateStop = this.convertValues(source["DateStop"], null);
	        this.Link = source["Link"];
	        this.Image = source["Image"];
	        this.Anzeigen = this.convertValues(source["Anzeigen"], sql.NullBool);
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
	export class Job {
	    ID: string;
	    Name: string;
	    Online: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Job(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Online = source["Online"];
	    }
	}
	export class Mitarbeiter {
	    ID: string;
	    Name: string;
	    Short: string;
	    Image: boolean;
	    Sex: string;
	    Tags: string;
	    Focus: string;
	    Abteilungid: string;
	
	    static createFrom(source: any = {}) {
	        return new Mitarbeiter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Short = source["Short"];
	        this.Image = source["Image"];
	        this.Sex = source["Sex"];
	        this.Tags = source["Tags"];
	        this.Focus = source["Focus"];
	        this.Abteilungid = source["Abteilungid"];
	    }
	}
	export class Partner {
	    ID: string;
	    Name: string;
	    Link: string;
	    Image: string;
	
	    static createFrom(source: any = {}) {
	        return new Partner(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Link = source["Link"];
	        this.Image = source["Image"];
	    }
	}

}

export namespace main {
	
	export class Counts {
	    Abteilung: number;
	    Angebote: number;
	    Jobs: number;
	    Mitarbeiter: number;
	    Partner: number;
	
	    static createFrom(source: any = {}) {
	        return new Counts(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Abteilung = source["Abteilung"];
	        this.Angebote = source["Angebote"];
	        this.Jobs = source["Jobs"];
	        this.Mitarbeiter = source["Mitarbeiter"];
	        this.Partner = source["Partner"];
	    }
	}
	export class Sg_Adressen {
	    SG_Adressen_PK: number;
	    Suchbegriff: sql.NullString;
	    KundNr: sql.NullString;
	    LiefNr: sql.NullString;
	    Homepage: sql.NullString;
	    Telefon1: sql.NullString;
	    Telefon2: sql.NullString;
	    Mobiltelefon1: sql.NullString;
	    Mobiltelefon2: sql.NullString;
	    EMail1: sql.NullString;
	    EMail2: sql.NullString;
	    KundUmsatz: sql.NullFloat64;
	    LiefUmsatz: sql.NullFloat64;
	
	    static createFrom(source: any = {}) {
	        return new Sg_Adressen(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SG_Adressen_PK = source["SG_Adressen_PK"];
	        this.Suchbegriff = this.convertValues(source["Suchbegriff"], sql.NullString);
	        this.KundNr = this.convertValues(source["KundNr"], sql.NullString);
	        this.LiefNr = this.convertValues(source["LiefNr"], sql.NullString);
	        this.Homepage = this.convertValues(source["Homepage"], sql.NullString);
	        this.Telefon1 = this.convertValues(source["Telefon1"], sql.NullString);
	        this.Telefon2 = this.convertValues(source["Telefon2"], sql.NullString);
	        this.Mobiltelefon1 = this.convertValues(source["Mobiltelefon1"], sql.NullString);
	        this.Mobiltelefon2 = this.convertValues(source["Mobiltelefon2"], sql.NullString);
	        this.EMail1 = this.convertValues(source["EMail1"], sql.NullString);
	        this.EMail2 = this.convertValues(source["EMail2"], sql.NullString);
	        this.KundUmsatz = this.convertValues(source["KundUmsatz"], sql.NullFloat64);
	        this.LiefUmsatz = this.convertValues(source["LiefUmsatz"], sql.NullFloat64);
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

export namespace sql {
	
	export class NullBool {
	    Bool: boolean;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NullBool(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Bool = source["Bool"];
	        this.Valid = source["Valid"];
	    }
	}
	export class NullFloat64 {
	    Float64: number;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NullFloat64(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Float64 = source["Float64"];
	        this.Valid = source["Valid"];
	    }
	}
	export class NullString {
	    String: string;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NullString(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.String = source["String"];
	        this.Valid = source["Valid"];
	    }
	}

}

