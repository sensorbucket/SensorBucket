import type {FeatureOfInterest} from "$lib/sensorbucket";
import {CSVParser} from "$lib/CSVParser";
import type {With} from "$lib/types";

// Partial feature
// Coded feature + Delete
type ContextFeatureOfInterest = Partial<FeatureOfInterest> & { delete?: boolean }
export type CSVFeatureOfInterest = With<ContextFeatureOfInterest, { name: string }>

function validateFeatureOfInterest(featureOfInterest: Partial<ContextFeatureOfInterest>): CSVFeatureOfInterest {
    if (featureOfInterest.name === undefined) {
        throw new Error("Feature of interest must have a name")
    }
    if (featureOfInterest.encoding_type !== undefined && featureOfInterest.encoding_type !== "application/geo+json") {
        throw new Error("Feature of interest must have encoding type application/geo+json")
    }
    return featureOfInterest as CSVFeatureOfInterest
}

interface Context {
    // Row specific
    featureOfInterest: ContextFeatureOfInterest
    // Global
    featuresOfInterest: CSVFeatureOfInterest[]
}

function createContext(): Context {
    return {
        featureOfInterest: {},
        featuresOfInterest: [],
    }
}

function assertGeometry(ctx: Context) {
    if (ctx.featureOfInterest.encoding_type === undefined && ctx.featureOfInterest.feature === undefined) {
        ctx.featureOfInterest.encoding_type = "application/geo+json"
        ctx.featureOfInterest.feature = {
            type: "Point",
            coordinates: [0, 0]
        }
    }
}

const parser = new CSVParser<Context>(createContext)
parser.addColumn(/^name$/i, (_) => (ctx, value) => {
    ctx.userData.featureOfInterest.name = value
})
parser.addColumn(/^description$/i, (_) => (ctx, value) => {
    ctx.userData.featureOfInterest.description = value
})
parser.addColumn(/^properties /i, (field) => (ctx, value) => {
    if (ctx.userData.featureOfInterest.properties === undefined) ctx.userData.featureOfInterest.properties = {}
    ctx.userData.featureOfInterest.properties[field.substring("properties ".length).replaceAll(" ", "__").toLowerCase()] = value
})
parser.addColumn(/^latitude$/i, (_) => (ctx, value) => {
    assertGeometry(ctx.userData)
    ctx.userData.featureOfInterest.feature.coordinates[0] = parseFloat(value)
})
parser.addColumn(/^longitude$/i, (_) => (ctx, value) => {
    assertGeometry(ctx.userData)
    ctx.userData.featureOfInterest.feature.coordinates[1] = parseFloat(value)
})
parser.addColumn(/^DELETE$/, (_) => (ctx, value) => {
    if (value !== "DELETE") return
    ctx.userData.featureOfInterest.delete = true;
})

parser.beforeRowParse = (context) => context.userData = {
    ...context.userData,
    featureOfInterest: {},
}
parser.afterRowParse = (context) => {
    const featureOfInterest = validateFeatureOfInterest(context.userData.featureOfInterest)
    context.userData.featuresOfInterest.push(featureOfInterest)
}
export const CSVFeatureOfInterestParser = parser;
