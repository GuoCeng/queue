package cron

type Option func(*Cron)

func WithMinutes() Option {
	return WithParser(NewParser(
		Minute | Hour | Dom | Month | Dow,
	))
}

func WithParser(p ScheduleParser) Option {
	return func(c *Cron) {
		c.parser = p
	}
}
