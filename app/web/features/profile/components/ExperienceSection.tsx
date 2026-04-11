import { useFieldArray, useFormContext } from "react-hook-form"
import { Profile } from "../schema"
import { InputField } from "@/components/ui/form/InputField"

export const ExperienceSection = () => {
  const { control } = useFormContext<Profile>()
  const { fields, append, remove } = useFieldArray({
    control,
    name: "experiences",
  })

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-bold">Experience</h3>
      {fields.map((field, index) => (
        <div
          key={field.id}
          className="relative rounded-md border bg-gray-50 p-4"
        >
          <InputField name={`experiences.${index}.title`} label="Job Title" />
          <InputField name={`experiences.${index}.company`} label="Company" />

          <button
            type="button"
            onClick={() => remove(index)}
            className="absolute top-2 right-2 text-sm text-red-500"
          >
            Remove
          </button>
        </div>
      ))}
      <button
        type="button"
        onClick={() => append({ title: "", company: "", startDate: "" })}
        className="font-medium text-blue-600 hover:underline"
      >
        + Add Experience
      </button>
    </div>
  )
}
