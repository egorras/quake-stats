namespace QuakeStats.Domain.Events;

public abstract record BaseEvent
{
    public Guid MatchGuid { get; init; }
    public int Time { get; init; }
    public bool Warmup { get; init; }
}
