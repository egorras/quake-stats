namespace QuakeStats.Domain.Entities;

public abstract class BaseEntity
{
    public int Id { get; set; }
    public DateTimeOffset CreatedAt { get; set; }
}
