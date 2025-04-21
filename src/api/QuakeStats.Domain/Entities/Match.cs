using QuakeStats.Domain.Enums;
using QuakeStats.Domain.Events;

namespace QuakeStats.Domain.Entities;

public class Match : BaseEntity
{
    public DateTimeOffset StartedAt { get; set; }
    public DateTimeOffset? ReportedAt { get; set; }
    public Guid MatchGuid { get; set; }
    public string Map { get; set; } = string.Empty;
    public GameType GameType { get; set; }
    public int TeamScoreRed { get; set; }
    public int TeamScoreBlue { get; set; }

    public void Apply(MatchStartedEvent @event)
    {
        MatchGuid = @event.MatchGuid;
        Map = @event.Map;
        GameType = @event.GameType;
        StartedAt = DateTimeOffset.UtcNow;
    }

    public void Apply(MatchReportEvent @event)
    {
        MatchGuid = @event.MatchGuid;
        Map = @event.Map;
        GameType = @event.GameType;
        TeamScoreRed = @event.TeamScoreRed;
        TeamScoreBlue = @event.TeamScoreBlue;
        ReportedAt = DateTimeOffset.UtcNow;
    }
}
