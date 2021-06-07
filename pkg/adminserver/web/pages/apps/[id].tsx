/* eslint-disable react/prop-types */
import { useRouter } from 'next/dist/client/router'
import React, { FC, useRef, useState } from 'react'
import { Heading, TextInput } from '../../src/components'
import { useGetFullAppQuery, useCreateReviewAppMutation, Tuple, useSetAppVarMutation, useSetAppEnvMutation, Exact, SetTupleInput, useDestroyAppMutation, useUnsetAppEnvMutation } from '../../src/generated/graphql'
import { throwError } from '../../src/throwError'
import { PencilAltIcon, PlusCircleIcon, SaveIcon, TrashIcon } from '@heroicons/react/outline'
import { UseMutationResponse } from 'urql'

const CreateForm = ({ app }) => {
	const [{ error, fetching }, create] = useCreateReviewAppMutation()
	const branchNameRef = useRef(null)

	const submit = (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault()

		create({
			input: {
				name:       app.name,
				branchName: branchNameRef.current.value,
			},
		})
	}

	return <form onSubmit={submit}>
		{error && <div className="bg-red-700 text-white border-red-700 p-2">
			{error.message}
		</div>}
		<TextInput name="branchName" ref={branchNameRef} placeholder="branch name" required disabled={fetching} />
		<button type="submit" className="rounded-sm p-1 underline" disabled={fetching}>Create</button>
	</form>
}

type AppVarFormProps = {
	appVar: Tuple
	defaultEdit?: boolean
	name: string
	finished?: () => void
	mutation: () => UseMutationResponse<any, Exact<{
		input: SetTupleInput
	}>>
}

const AppVarForm: FC<AppVarFormProps> = ({ name, appVar, defaultEdit = false, finished, mutation }) => {
	const [editing, setEditing] = useState<boolean>(defaultEdit)
	const [{ error }, save] = mutation()
	const keyRef = useRef(null)
	const valueRef = useRef(null)
	const formRef = useRef(null)

	const submit = async (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault()

		const result = await save({
			input: {
				name,
				key:   keyRef.current.value,
				value: valueRef.current.value,
			},
		})

		if (!result.error) {
			setEditing(false)
			finished && finished()
		}
	}

	return <form ref={formRef} onSubmit={submit}>
		{error && <div className="bg-red-700 text-white border-red-700 p-2">
			{error.message}
		</div>}

		<TextInput
			name="key"
			disabled={!editing || !defaultEdit}
			required
			ref={keyRef}
			defaultValue={appVar.key}
			placeholder="key"
		/>

		<TextInput
			name="value"
			disabled={!editing}
			required
			ref={valueRef}
			defaultValue={appVar.value}
			placeholder="value"
		/>

		{editing
			? <button type="submit"><SaveIcon className="w-5" /></button>
			: <PencilAltIcon className="w-5" onClick={() => setEditing(true)} />}
	</form>
}

const ShowApp = () => {
	const router = useRouter()
	const id = router.query.id as string
	const [{ data: { getApp: app } }] = throwError(useGetFullAppQuery({ variables: { name: id } }))
	const [{ error: destroyAppError }, destroyApp] = useDestroyAppMutation()
	const [{ error: unsetAppVarError }, unsetAppEnv] = useUnsetAppEnvMutation()
	const hostname = `https://${app.name}.staging.freshlyservices.net/`
	const [addNew, setAddNew] = useState<boolean>(false)

	return <div>
		<div className="border-b pb-2 mb-2">
			<Heading>{app.name}</Heading>

			<p>
				Available at - <a href={hostname}>{hostname}</a> - if it uses your cluster&apos;s default hostname.
			</p>
		</div>

		{app.reviewApp || <>
			<div className="border-b pb-2 mb-2">
				<Heading>Create a review app</Heading>
				<CreateForm app={app} />
				<Heading>Review apps</Heading>
				{destroyAppError && <div className="bg-red-700 text-white border-red-700 p-2">
					{destroyAppError.message}
				</div>}

				{app.reviewApps && app.reviewApps.map(reviewApp =>
					<div key={reviewApp.name}>
						<a href={`/tuber/apps/${reviewApp.name}`}>{reviewApp.name}</a>
						<TrashIcon className="w-5" onClick={() => destroyApp({ input: { name: reviewApp.name } })}/>
					</div>,
				)}
			</div>
		</>}

		<div className="border-b pb-2 mb-2">
			<Heading>YAML Interpolation Vars</Heading>
			{app.vars.map(appVar => <AppVarForm key={appVar.key} name={app.name} appVar={appVar} mutation={useSetAppVarMutation} />)}

			{addNew
				? <AppVarForm name={app.name} appVar={{} as Tuple} defaultEdit finished={() => setAddNew(false)} mutation={useSetAppVarMutation} />
				: <PlusCircleIcon className="w-5" onClick={() => setAddNew(true)} />}
		</div>

		<div className="border-b pb-2 mb-2">
			<Heading> Environment Variables </Heading>
			{app.env.map(appVar =>
				<div key={appVar.key}>
					<AppVarForm key={appVar.key} name={app.name} appVar={appVar} mutation={useSetAppEnvMutation} />
					<TrashIcon className="w-5" onClick={() => unsetAppEnv({ input: { name: app.name, key: appVar.key, value: appVar.value } })}/>
				</div>,
			)}

			{addNew
				? <AppVarForm name={app.name} appVar={{} as Tuple} defaultEdit finished={() => setAddNew(false)} mutation={useSetAppEnvMutation} />
				: <PlusCircleIcon className="w-5" onClick={() => setAddNew(true)} />}
		</div>
	</div>
}

export default ShowApp
