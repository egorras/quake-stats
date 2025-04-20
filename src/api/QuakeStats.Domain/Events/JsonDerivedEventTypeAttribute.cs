using System.Text.Json.Serialization;

namespace QuakeStats.Domain.Events;

/// <summary>
/// Attribute to automatically register derived event types with JsonDerivedType using the event's EventType constant
/// </summary>
/// <typeparam name="T">The event type to register</typeparam>
public class JsonDerivedEventTypeAttribute<T> : JsonDerivedTypeAttribute where T : BaseEvent
{
    public JsonDerivedEventTypeAttribute() : base(typeof(T), GetEventType())
    {
    }

    private static string GetEventType()
    {
        var property = typeof(T).GetProperty(nameof(BaseEvent.EventType));
        
        if (property?.GetMethod?.IsStatic == true)
        {
            return property.GetValue(null) as string ?? string.Empty;
        }
        
        var field = typeof(T).GetField(nameof(BaseEvent.EventType));
        return field?.GetValue(null) as string ?? string.Empty;
    }
}
