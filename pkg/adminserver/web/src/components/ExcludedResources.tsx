import React, { FC, useState, useRef } from 'react'
import { TextInput } from '../../src/components'
import { SetResourceInput, Exact, Resource } from '../generated/graphql'
import { UseMutationResponse } from 'urql'
import { AddButton } from './AddButton'
import { SaveIcon, XCircleIcon, TrashIcon } from '@heroicons/react/outline'

type Props = {
	appName: string
	resources: Pick<Resource, | 'kind' | 'name'>[]
	useSet: () => UseMutationResponse<any, Exact<{ input: SetResourceInput }>>
	useUnset: () => UseMutationResponse<any, Exact<{ input: SetResourceInput }>>
}

export const ExcludedResources:FC<Props> = ({ appName, resources, useSet, useUnset }) => {
	const [addNew, setAddNew] = useState<boolean>(false)
	const nameRef = useRef(null)
	const kindRef = useRef(null)

	const [{ error: setErr }, set] = useSet()
	const [{ error: unsetErr }, unset] = useUnset()

	const err = setErr || unsetErr

	const doSet = async (event) => {
		event.preventDefault()

		await set({
			input: {
				appName: appName,
				name:    nameRef.current.value,
				kind:    kindRef.current.value,
			},
		})
	}

	const doUnset = resource => async (event) => {
		event.preventDefault()

		await unset({
			input: {
				appName: appName,
				name:    resource.name,
				kind:    resource.kind,
			},
		})
	}

	return <div>
		{resources.map(resource =>
			<div key={resource.name} className="pb-1">
				<span>{resource.name}</span>
				<span className="pl-3 pr-1">{resource.kind}</span>
				<TrashIcon className="w-5 text-red-600" onClick={doUnset(resource)} />
			</div>,
		)}

		{err && <div className="bg-red-700 text-white border-red-700 p-2">
			{err.message}
		</div>}

		{addNew &&
			<form className="inline" onSubmit={doSet}>
				<label>Name</label>
				<TextInput required ref={nameRef} />
				<label>Kind</label>
				<TextInput required ref={kindRef} />
				<XCircleIcon className="w-5 select-none" onClick={() => { setAddNew(false) }} />
				<button><SaveIcon className="w-5" /></button>
			</form>}

		{addNew ||
			<AddButton onClick={() => setAddNew(true)} />}
	</div>
}