using QuakeStats.Domain.Entities;

namespace QuakeStats.Domain.Services;

public interface IEventProcessor
{
    Task ProcessEventAsync(Event @event);
}