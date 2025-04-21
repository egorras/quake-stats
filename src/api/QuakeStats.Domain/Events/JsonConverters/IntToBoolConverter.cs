using System.Text.Json;
using System.Text.Json.Serialization;

namespace QuakeStats.Domain.Events.JsonConverters;

public class IntToBoolConverter : JsonConverter<bool>
{
    public override bool Read(ref Utf8JsonReader reader, Type typeToConvert, JsonSerializerOptions options)
    {
        if (reader.TokenType == JsonTokenType.Number)
        {
            var value = reader.GetInt32();
            return value == 1;
        }
        return reader.GetBoolean();
    }

    public override void Write(Utf8JsonWriter writer, bool value, JsonSerializerOptions options)
    {
        throw new NotImplementedException();
    }
}
