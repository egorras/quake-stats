using QuakeStats.Domain.Events;

namespace QuakeStats.Domain.Entities;

public class Match : BaseEntity
{
    public Guid MatchGuid { get; set; }
    public required string Map { get; set; }

    public void Apply(MatchStartedEvent @event)
    {
        MatchGuid = @event.MatchGuid;
        Map = @event.Map;
    }
}
