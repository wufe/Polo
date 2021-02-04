import { IApplicationBranchModel } from '@/state/models';
import { values } from 'mobx';
import { observer } from 'mobx-react-lite';
import React from 'react';
import dayjs from 'dayjs';

type TProps = {
    branches: IApplicationBranchModel[];
    onSessionCreationSubmission: (checkout: string) => void;
}

export const ApplicationBranches = observer((props: TProps) => {
    return <>
        <h4 className="mt-2 mb-1 text-sm text-gray-500 uppercase">Branches:</h4>
        <div
            className="grid items-end gap-2" style={{ gridTemplateColumns: '2fr 3fr minmax(250px, 2fr) minmax(150px, 1fr) minmax(150px, 1fr)', gridTemplateRows: '3m'}}>
            {sortBranches(props.branches).map((branch, key) =>
            <React.Fragment key={key}>
                <span className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis " title={branch.name}>{branch.name}</span>
                <div className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis ">{branch.message}</div>
                <div className="">
                    <div className="text-xs text-gray-500 uppercase">Last author</div>
                    <div className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis">{branch.author}</div>
                </div>
                <div className="">
                    <div className="text-xs text-gray-500 uppercase">Updated</div>
                    <div className="text-sm whitespace-nowrap">{dayjs(branch.date).format('DD MMM HH:mm')}</div>
                </div>
                <span className=" text-center">
                    <span className="text-sm underline cursor-pointer inline-block mx-3 hover:text-blue-400" onClick={() => props.onSessionCreationSubmission(branch.name)}>Create session</span>
                </span>
            </React.Fragment>)}
        </div>
    </>;
});

function sortBranches(branches: IApplicationBranchModel[]): IApplicationBranchModel[] {
    const preferred = [
        'master',
        'main',
        'hotfix',
        'develop',
        'dev',
        'feature'
    ];

    let result: IApplicationBranchModel[] = [];
    let length = branches.length;

    for (const pref of preferred) {
        for (let i = 0; i < length; i++) {
            const branch = branches[i];
            if (branch.name.toLowerCase().startsWith(pref.toLowerCase())) {
                result.push(branch);
                branches.splice(i, 1)
                length--;
                i--;
            }
        }
    }

    return result.concat(branches);
}