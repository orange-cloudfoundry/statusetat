# golang-ical
A  ICS / ICal parser and serialiser for Golang.

[![GoDoc](https://godoc.org/github.com/arran4/golang-ical?status.svg)](https://godoc.org/github.com/arran4/golang-ical)

Because the other libraries didn't quite do what I needed.

Usage, parsing:
```golang
    cal, err := ParseCalendar(strings.NewReader(input))

```

Usage, parsing from a URL:
```golang
    cal, err := ParseCalendarFromUrl("https://your-ics-url")
```

Creating:
```golang
  cal := ics.NewCalendar()
  cal.SetMethod(ics.MethodRequest)
  event := cal.AddEvent(fmt.Sprintf("id@domain", p.SessionKey.IntID()))
  event.SetCreatedTime(time.Now())
  event.SetDtStampTime(time.Now())
  event.SetModifiedAt(time.Now())
  event.SetStartAt(time.Now())
  event.SetEndAt(time.Now())
  event.SetSummary("Summary")
  event.SetLocation("Address")
  event.SetDescription("Description")
  event.SetURL("https://URL/")
  event.AddRrule(fmt.Sprintf("FREQ=YEARLY;BYMONTH=%d;BYMONTHDAY=%d", time.Now().Month(), time.Now().Day()))
  event.SetOrganizer("sender@domain", ics.WithCN("This Machine"))
  event.AddAttendee("reciever or participant", ics.CalendarUserTypeIndividual, ics.ParticipationStatusNeedsAction, ics.ParticipationRoleReqParticipant, ics.WithRSVP(true))
  return cal.Serialize()
```

Helper methods created as needed feel free to send a P.R. with more.

# Working with Recurring Events

This library parses and provides typed access to recurrence properties (`RRULE`, `RDATE`, `EXDATE`, `EXRULE`, `RECURRENCE-ID`) but does not expand them into concrete occurrence dates.
This keeps the library dependency-free and lets callers choose their own expansion strategy.

## Accessing recurrence properties

```golang
  // Parse RRULE into a structured type
  rules, err := event.GetRRules()
  for _, rule := range rules {
      fmt.Println(rule.Freq)       // e.g. ics.FrequencyYearly
      fmt.Println(rule.ByMonth)    // e.g. [10]
      fmt.Println(rule.ByDay)      // e.g. [{OrdWeek:-1 Day:SU}]
      fmt.Println(rule.Count)      // e.g. 10
      fmt.Println(rule.Interval)   // e.g. 1
  }

  // Get excluded dates (handles comma-separated values and multiple properties)
  exDates, err := event.GetExDates()    // []time.Time
  rDates, err := event.GetRDates()      // []time.Time

  // Get recurrence ID (for modified/cancelled occurrences)
  recID, err := event.GetRecurrenceID() // time.Time

  // Serialize a RecurrenceRule back to RRULE format
  rule, _ := ics.ParseRecurrenceRule("FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE,FR")
  fmt.Println(rule.String()) // "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE,FR"
```

## Expanding occurrences with rrule-go

To expand recurring events into concrete dates, use a library like [rrule-go](https://github.com/teambition/rrule-go):

```golang
  import (
      "github.com/arran4/golang-ical"
      "github.com/teambition/rrule-go"
  )

  // Parse your calendar
  cal, _ := ics.ParseCalendar(reader)

  for _, event := range cal.Events() {
      dtstart, _ := event.GetStartAt()
      dtend, _ := event.GetEndAt()
      duration := dtend.Sub(dtstart)

      // Build an RRuleSet from the event's recurrence properties
      set := &rrule.Set{}

      rules, _ := event.GetRRules()
      for _, r := range rules {
          rr, _ := rrule.StrToRRule(r.String())
          rr.DTStart(dtstart)
          set.RRule(rr)
      }

      rDates, _ := event.GetRDates()
      for _, rd := range rDates {
          set.RDate(rd)
      }

      exDates, _ := event.GetExDates()
      for _, exd := range exDates {
          set.ExDate(exd)
      }

      // Get all occurrences in a time range
      occurrences := set.Between(rangeStart, rangeEnd, true)

      // Each occurrence is a start time; compute the end from the original duration
      for _, occ := range occurrences {
          fmt.Printf("  %s to %s\n", occ, occ.Add(duration))
      }
  }
```

### Handling RECURRENCE-ID overrides

Calendar feeds use `RECURRENCE-ID` to modify or cancel individual occurrences of a recurring event.
These appear as separate VEVENTs with the same UID but a `RECURRENCE-ID` property indicating which occurrence they replace.

```golang
  type overrideKey struct {
      UID   string
      Start time.Time
  }

  overrides := map[overrideKey]*ics.VEvent{}
  var regularEvents []*ics.VEvent

  for _, event := range cal.Events() {
      recID, err := event.GetRecurrenceID()
      if err == nil {
          uid := event.GetProperty(ics.ComponentPropertyUniqueId).Value
          overrides[overrideKey{UID: uid, Start: recID}] = event
      } else {
          regularEvents = append(regularEvents, event)
      }
  }

  // After expanding occurrences from regularEvents, check each generated
  // occurrence against the overrides map. If a match is found, replace that
  // occurrence with the override's times/properties (or remove it if the
  // override represents a cancellation).
```

# Notice

Looking for a co-maintainer.
