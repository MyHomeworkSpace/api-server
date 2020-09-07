package cornell

type ClassItem struct {
	CourseID          int    `json:"crseId"`
	CourseOfferNumber int    `json:"crseOfferNbr"`
	Subject           string `json:"subject"`
	CatalogNumber     string `json:"catalogNbr"`
	Title             string `json:"title"`
	ClassNumbers      []int  `json:"classNbrs"`
	Units             int    `json:"units"`
	SchedulePrint     bool   `json:"schedulePrint"`
	Exists            bool   `json:"exists"`
}

type CoursePairs struct {
	CoursePairs []string `json:"crsePairs"`
}

type Course struct {
	Strm              int           `json:"strm"`
	CourseID          int           `json:"crseId"`
	CourseOfferNumber int           `json:"crseOfferNbr"`
	Subject           string        `json:"subject"`
	CatalogNumber     string        `json:"catalogNbr"`
	TitleShort        string        `json:"titleShort"`
	TitleLong         string        `json:"titleLong"`
	EnrollGroups      []EnrollGroup `json:"enrollGroups"`
	Cap               string        `json:"capDttm"`
	RosterPrint       bool          `json:"rosterPrint"`
}

type EnrollGroup struct {
	ClassSections             []Section   `json:"classSections"`
	UnitsMinimum              int         `json:"unitsMinimum"`
	UnitsMaximum              int         `json:"unitsMaximum"`
	ComponentsOptional        []string    `json:"componentsOptional"`
	ComponentsRequired        []string    `json:"componentsRequired"`
	SessionCode               string      `json:"sessionCode"`
	SessionBegin              string      `json:"sessionBeginDt"`
	SessionEnd                string      `json:"sessionEndDt"`
	Session                   string      `json:"sessionLong"`
	SyllabusReferenceMap      interface{} `json:"syllabusReferenceMap"` //no idea what this is meant to be user for
	SyllabusReferenceMapCount int         `json:"syllabusReferenceMapCount"`
	SyllabusPublishedMapCount int         `json:"syllabusPublishedMapCount"`
	Syllabuses                []Syllabus  `json:"syllabusReferences"`
	RosterPrint               bool        `json:"rosterPrint"`
}

type Section struct {
	Component              string      `json:"ssrComponent"`
	ComponentLong          string      `json:"ssrComponentLong"`
	Section                string      `json:"section"`
	ClassNum               int         `json:"classNbr"`
	Meetings               []Meeting   `json:"meetings"`
	Campus                 string      `json:"campus"`
	CampusDesc             string      `json:"campusDescr"`
	AcadOrg                string      `json:"acadOrg"`
	Location               string      `json:"location"`
	LocationDesc           string      `json:"locationDescr"`
	Start                  string      `json:"startDt"`
	End                    string      `json:"endDt"`
	AddConsent             string      `json:"addConsent"`
	AddConsentDescr        string      `json:"addConsentDescr"`
	ComponentGraded        bool        `json:"isComponentGraded"`
	InstructionMode        string      `json:"instructionMode"`
	InstrModeDescShort     string      `json:"instrModeDescrshort"`
	InstrModeDesc          string      `json:"instrModeDescr"`
	TopicDesc              string      `json:"topicDesc"`
	CombinedSkipMtgpatEdit interface{} `json:"combinedSkipMtgpatEdit"` //i have no idea what this is
	OpenStatus             string      `json:"openStatus"`
	OpenStatusDesc         string      `json:"openStatusDescr"`
	RosterPrint            bool        `json:"rosterPrint"`
}

type Meeting struct {
	Number            int    `json:"classMtgNumber"`
	StartTime         string `json:"timeStart"`
	EndTime           string `json:"timeEnd"`
	Mon               string `json:"mon"` //for some reason beyond me these are stored as strings.
	Tue               string `json:"tue"`
	Wed               string `json:"wed"`
	Thu               string `json:"thu"`
	Fri               string `json:"fri"`
	Sat               string `json:"sat"`
	Sun               string `json:"sun"`
	StartDate         string `json:"startDt"`
	EndDate           string `json:"endDt"`
	Pattern           string `json:"pattern"`
	FacilityDesc      string `json:"facilityDescr"`
	BuildingDesc      string `json:"buildingDescr"`
	FacilityDescShort string `json:"facilityDescrshort"`
	MeetingTopicDesc  string `json:"meetingTopicDescription"`
}

type Syllabus struct {
	RosterSlug      string `json:"rosterSlug"`
	LinkID          string `json:"linkId"`
	SyllabusID      string `json:"syllabusId"`
	Type            string `json:"type"`
	Version         string `json:"version"`
	ViewPermission  string `json:"viewPermission"`
	Updated         string `json:"updatedDttm"`
	Published       string `json:"publishedDttm"`
	ResourceID      string `json:"resourceId"`
	ResourceAdded   string `json:"resourceAddedDttm"`
	ResourceUpdated string `json:"resourceUpdatedDttm"`
}

type Class struct {
	Subject       string // this is the subject, e.g. in CHEM 2090, it would be CHEM
	CatalogNum    string // this is the course number, e.g. in CHEM 2090, it would be 2090
	ClassNum      string // this is the class number, e.g. in CHEM 2090, it might be 4749
	Component     string // this would be LEC for example
	ComponentLong string // this would be Lecture for example
	Section       string // this is the section number, like 403, not to be confused with the class number
	Campus        string // this is the campus, for example MAIN or TECH
	CampusLong    string // this is the long name for the campus, for example, Ithaca
	Location      string // this is the abbreviated name of the campus, for example, ITH
	LocationLong  string // this is the long name of the campus, for example, "Ithaca, NY (Main Campus)"
}
