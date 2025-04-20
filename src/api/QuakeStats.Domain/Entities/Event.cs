namespace QuakeStats.Domain.Entities;

public class Event : BaseEntity
{
    public required string Type { get; set; }
    public required string Data { get; set; }
}
