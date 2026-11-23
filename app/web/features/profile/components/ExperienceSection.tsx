"use client"

import { useFieldArray, useFormContext } from "react-hook-form"
import type { ProfileType } from "../schemas"
import { InputField } from "@/components/ui/form/InputField"

export const ExperienceSection = () => {
  const { control } = useFormContext<ProfileType>()
  const { fields, append, remove } = useFieldArray({
    control,
    name: "experiences",
  })

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-bold">Experience</h3>
      {fields.map((field, index) => (
        <div key={field.id} className="relative space-y-2 rounded-md border bg-gray-50 p-4">
          <InputField name={`experiences.${index}.title`} label="Job title" />
          <InputField name={`experiences.${index}.company`} label="Company" />
          <InputField name={`experiences.${index}.startDate`} label="Start date" />
          <InputField name={`experiences.${index}.endDate`} label="End date" />
          <InputField name={`experiences.${index}.description`} label="Description" />

          <button type="button" onClick={() => remove(index)} className="text-sm text-red-500">
            Remove
          </button>
        </div>
      ))}
      <button
        type="button"
        onClick={() => append({ title: "", company: "", startDate: "", endDate: null, description: "" })}
        className="font-medium text-blue-600 hover:underline"
      >
        + Add experience
      </button>
      <p className="text-xs text-gray-500">
        Experiences can be edited in this form, but the current GraphQL UpdateProfileInput does not accept an experiences field, so this section cannot persist changes through the supplied API schema.
      </p>
    </div>
  )
}
