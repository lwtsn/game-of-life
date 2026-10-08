import { create } from '@bufbuild/protobuf'
import { RulesSchema } from '../gen/life/v1/rules_pb.js'

const colourPattern = new RegExp(create(RulesSchema).colourPattern)

export function isHex(value: unknown): value is string {
  return typeof value === 'string' && colourPattern.test(value)
}
