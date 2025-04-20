namespace QuakeStats.Domain.Entities;

public class Event : BaseEntity
{
    public required string EventType { get; set; }
    public required string EventData { get; set; }
    public bool Processed { get; set; } = false;
}
